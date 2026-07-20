package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"noant/config"
	"noant/internal/infrastructure"
)

// ========== MEDIA TYPES ==========

const (
	MediaTypeImage    = "image"
	MediaTypeDocument = "document"
	MediaTypeAudio    = "audio"
	MediaTypeVideo    = "video"
	MediaTypeSticker  = "sticker"
	MediaTypeLocation = "location"
	MediaTypeContact  = "contact"

	MaxThumbnailWidth = 400
)

var supportedImageFormats = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

// ========== MEDIA HANDLER ==========

type MediaHandler struct {
	cfg        *config.Config
	openwa     *OpenWAService
	redis      *infrastructure.RedisClient
	logger     *infrastructure.Logger
	httpClient *http.Client
}

func NewMediaHandler(cfg *config.Config, openwa *OpenWAService, redis *infrastructure.RedisClient, logger *infrastructure.Logger) *MediaHandler {
	return &MediaHandler{
		cfg:    cfg,
		openwa: openwa,
		redis:  redis,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (mh *MediaHandler) EnsureMediaDir() error {
	return os.MkdirAll(mh.cfg.OpenWAMediaDir, 0o750)
}

// DownloadMedia downloads a file from OpenWA media URL and stores it locally
func (mh *MediaHandler) DownloadMedia(ctx context.Context, sessionID, mediaURL, mediaType, mimeType string) (filePath, thumbPath string, fileSize int64, err error) {
	if mediaURL == "" {
		return "", "", 0, fmt.Errorf("media URL is empty")
	}

	if err := mh.EnsureMediaDir(); err != nil {
		return "", "", 0, fmt.Errorf("failed to create media dir: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", mediaURL, http.NoBody)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to create download request: %w", err)
	}

	if mh.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", mh.cfg.OpenWAApiKey)
	}

	resp, err := mh.httpClient.Do(req)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to download media: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", "", 0, fmt.Errorf("media download returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to read media data: %w", err)
	}

	fileSize = int64(len(data))

	ext := guessExtension(mimeType, mediaType)
	filename := fmt.Sprintf("%s_%d%s", sessionID, time.Now().UnixNano(), ext)
	filePath = filepath.Join(mh.cfg.OpenWAMediaDir, filename)

	if err := os.WriteFile(filePath, data, 0o640); err != nil {
		return "", "", 0, fmt.Errorf("failed to write media file: %w", err)
	}

	if mediaType == MediaTypeImage && isImageMime(mimeType) {
		thumbPath, err = generateThumbnail(filePath, filename)
		if err != nil {
			mh.logger.Warn("Failed to generate thumbnail", "error", err)
		}
	}

	mh.logger.Info("Media downloaded", "path", filePath, "size", fileSize, "type", mediaType)
	return filePath, thumbPath, fileSize, nil
}

func (mh *MediaHandler) HandleIncomingMedia(ctx context.Context, sessionID, userID string, media *OpenWAMediaData) (filePath, thumbPath *string, err error) {
	if !media.HasMedia || media.MediaURL == "" {
		return nil, nil, fmt.Errorf("no media data in message")
	}

	fp, tp, _, err := mh.DownloadMedia(ctx, sessionID, media.MediaURL, media.MediaType, media.MimeType)
	if err != nil {
		mh.logger.Error("Failed to download incoming media", "error", err, "mediaType", media.MediaType)
		return nil, nil, err
	}

	mh.logger.Info("Incoming media processed", "type", media.MediaType, "mime", media.MimeType, "path", fp)
	return &fp, &tp, nil
}

func (mh *MediaHandler) SendMediaFromUpload(ctx context.Context, sessionID, chatID, filePath, caption string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read media file: %w", err)
	}

	ext := filepath.Ext(filePath)
	mimeType := mimeTypeByExtension(ext)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))

	return mh.openwa.SendMediaMessage(sessionID, chatID, dataURL, caption)
}

func (mh *MediaHandler) CleanupExpiredMedia() (int64, error) {
	cutoff := time.Now().Add(-mh.cfg.OpenWAMediaRetention)
	var count int64

	err := filepath.Walk(mh.cfg.OpenWAMediaDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if info.ModTime().Before(cutoff) {
			if removeErr := os.Remove(path); removeErr == nil {
				count++
			}
		}
		return nil
	})

	return count, err
}

// ========== HELPER FUNCTIONS ==========

func guessExtension(mimeType, mediaType string) string {
	if mimeType != "" {
		switch mimeType {
		case "image/jpeg":
			return ".jpg"
		case "image/png":
			return ".png"
		case "image/webp":
			return ".webp"
		case "image/gif":
			return ".gif"
		case "application/pdf":
			return ".pdf"
		case "audio/ogg":
			return ".ogg"
		case "audio/mpeg", "audio/mp3":
			return ".mp3"
		case "audio/wav":
			return ".wav"
		case "video/mp4":
			return ".mp4"
		case "video/3gpp":
			return ".3gp"
		case "text/csv":
			return ".csv"
		}
	}

	switch mediaType {
	case MediaTypeImage:
		return ".jpg"
	case MediaTypeDocument:
		return ".bin"
	case MediaTypeAudio:
		return ".ogg"
	case MediaTypeVideo:
		return ".mp4"
	case MediaTypeSticker:
		return ".webp"
	default:
		return ".bin"
	}
}

func isImageMime(mimeType string) bool {
	return supportedImageFormats[mimeType]
}

func generateThumbnail(filePath, filename string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	img, _, err := image.Decode(f)
	if err != nil {
		return "", fmt.Errorf("failed to decode image for thumbnail: %w", err)
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	if w <= MaxThumbnailWidth {
		return "", nil
	}

	thumbFilename := "thumb_" + filename
	thumbPath := filepath.Join(filepath.Dir(filePath), thumbFilename)

	// Simple nearest-neighbor resize
	h := bounds.Dy()
	newW := MaxThumbnailWidth
	newH := h * MaxThumbnailWidth / w

	thumb := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			sx := x * w / newW
			sy := y * h / newH
			thumb.Set(x, y, img.At(sx, sy))
		}
	}

	outF, err := os.Create(thumbPath)
	if err != nil {
		return "", err
	}
	defer func() { _ = outF.Close() }()

	if err := png.Encode(outF, thumb); err != nil {
		return "", err
	}

	return thumbPath, nil
}

// OpenWAMediaData holds media-specific webhook data from OpenWA
type OpenWAMediaData struct {
	ID        string      `json:"id"`
	From      string      `json:"from"`
	To        string      `json:"to"`
	Type      string      `json:"type"`
	Timestamp interface{} `json:"timestamp"`
	FromMe    bool        `json:"fromMe"`
	Body      string      `json:"body"`
	HasMedia  bool        `json:"hasMedia"`
	MediaType string      `json:"mediaType"`
	MimeType  string      `json:"mimeType"`
	FileName  string      `json:"fileName"`
	FileSize  int64       `json:"fileSize"`
	MediaURL  string      `json:"mediaUrl"`
	Width     int         `json:"width"`
	Height    int         `json:"height"`
	Duration  int         `json:"duration"`
	Latitude  float64     `json:"latitude"`
	Longitude float64     `json:"longitude"`
	Address   string      `json:"address"`
	VCard     string      `json:"vcard"`
	Sender    OpenWASender `json:"sender"`
}

func mimeTypeByExtension(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".mp3":
		return "audio/mpeg"
	case ".ogg":
		return "audio/ogg"
	case ".wav":
		return "audio/wav"
	case ".mp4":
		return "video/mp4"
	case ".csv":
		return "text/csv"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}
