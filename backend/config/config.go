package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	JWTSecret     string
	SessionSecret string
	NodeEnv       string
	LogLevel      string
	APIURL        string
	AppURL        string

	// Cache
	CacheTTL     time.Duration
	CacheMaxKeys int

	// TiDB
	TiDBHost     string
	TiDBPort     int
	TiDBUser     string
	TiDBPassword string
	TiDBDatabase string
	DBPoolSize   int

	// Redis
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisSSL      bool
	RedisShortTTL time.Duration

	// Groq
	GroqAPIKeys []string

	// Twilio
	TwilioAccountSID     string
	TwilioAuthToken      string
	TwilioWhatsAppNumber string

	// Resend
	ResendAPIKey string
	ResendFrom   string

	// SMTP / Gmail
	SMTPHost       string
	SMTPPort       int
	SMTPUsername   string
	SMTPPassword   string
	SMTPFrom       string
	SMTPSkipVerify bool

	// Polar
	PolarAccessToken    string
	PolarOrganizationID string
	PolarServerURL      string
	PolarWebhookSecret  string
	// Billing system - Polar checkout URLs
	PolarPulseSmallURL  string
	PolarPulseMediumURL string
	PolarPulseLargeURL  string
	PolarProMonthlyURL  string
	PolarProAnnualURL   string
	PolarEnterpriseURL  string
	CORSOrigins         []string

	// Telegram
	TelegramBotToken   string
	TelegramWebhookURL string

	// Meta (WhatsApp / Facebook / Instagram)
	MetaAccessToken    string
	MetaPhoneNumberID  string
	MetaPageID         string
	InstagramAccountID string
	MetaVerifyToken    string

	// Web Push (VAPID)
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string

	// OpenWA (self-hosted WhatsApp API)
	OpenWAEnabled          bool
	OpenWABaseURL          string
	OpenWAApiKey           string
	OpenWASessionID        string
	OpenWAWebhookSecret    string
	OpenWARateLimitText    int
	OpenWARateLimitMedia   int
	OpenWARateLimitTemplate int
	OpenWARateLimitBurst   int
	OpenWAQueueDepth       int
	OpenWAMediaDir         string
	OpenWAMediaRetention   time.Duration
	OpenWASessionHealthInterval time.Duration
	OpenWAMaxReconnectAttempts   int
	OpenWAConnPoolSize     int
	OpenWAConnTimeout      time.Duration
	OpenWAReqTimeout       time.Duration
}

func Load() *Config {
	paths := []string{".env", "../.env", "../../.env"}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			_ = godotenv.Load(p)
			break
		}
	}
	port := getEnv("PORT", "5000")

	cacheTTL, _ := strconv.Atoi(getEnv("CACHE_TTL", "300"))
	cacheMaxKeys, _ := strconv.Atoi(getEnv("CACHE_MAX_KEYS", "10000"))
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	tidbPort, _ := strconv.Atoi(getEnv("TIDB_PORT", "4000"))
	dbPoolSize, _ := strconv.Atoi(getEnv("DB_POOL_SIZE", "200"))

	redisPort, _ := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	redisSSL, _ := strconv.ParseBool(getEnv("REDIS_SSL", "true"))
	redisShortTTL, _ := strconv.Atoi(getEnv("REDIS_SHORT_TTL", "259200"))

	var groqKeys []string
	if raw := strings.TrimSpace(os.Getenv("GROQ_API_KEY")); raw != "" {
		for _, key := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(key); trimmed != "" {
				groqKeys = append(groqKeys, trimmed)
			}
		}
	}
	for i := 1; i <= 10; i++ {
		key := os.Getenv(fmt.Sprintf("GROQ_API_KEY_%d", i))
		if key != "" {
			groqKeys = append(groqKeys, key)
		}
	}

	cfg := &Config{
		Port:          port,
		JWTSecret:     getEnv("JWT_SECRET", ""),
		SessionSecret: getEnv("SESSION_SECRET", ""),
		NodeEnv:       getEnv("NODE_ENV", "development"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		APIURL:        getEnv("API_URL", "http://localhost:"+port),
		AppURL:        getEnv("APP_URL", "http://localhost:3000"),
		VAPIDPublicKey:  getEnv("VAPID_PUBLIC_KEY", ""),
		VAPIDPrivateKey: getEnv("VAPID_PRIVATE_KEY", ""),
		VAPIDSubject:    getEnv("VAPID_SUBJECT", "mailto:support@noant.io"),

		CacheTTL:     time.Duration(cacheTTL) * time.Second,
		CacheMaxKeys: cacheMaxKeys,

		TiDBHost:     getEnv("TIDB_HOST", ""),
		TiDBPort:     tidbPort,
		TiDBUser:     getEnv("TIDB_USER", ""),
		TiDBPassword: getEnv("TIDB_PASSWORD", ""),
		TiDBDatabase: getEnv("TIDB_DATABASE", "noant"),
		DBPoolSize:   dbPoolSize,

		RedisHost:     getEnv("REDIS_HOST", ""),
		RedisPort:     redisPort,
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisSSL:      redisSSL,
		RedisShortTTL: time.Duration(redisShortTTL) * time.Second,

		GroqAPIKeys: groqKeys,

		TwilioAccountSID:     getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:      getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioWhatsAppNumber: getEnv("TWILIO_WHATSAPP_NUMBER", ""),

		ResendAPIKey: getEnv("RESEND_API_KEY", ""),
		ResendFrom:   getEnv("RESEND_FROM", "onboarding@resend.dev"),

		SMTPHost:       getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:       smtpPort,
		SMTPUsername:   getEnv("SMTP_USERNAME", ""),
		SMTPPassword:   getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:       getEnv("SMTP_FROM", ""),
		SMTPSkipVerify: getEnv("SMTP_SKIP_VERIFY", "false") == "true",

		PolarAccessToken:    getEnv("POLAR_ACCESS_TOKEN", ""),
		PolarOrganizationID: getEnv("POLAR_ORGANIZATION_ID", ""),
		PolarServerURL:      getEnv("POLAR_SERVER_URL", "https://api.polar.sh"),
		PolarWebhookSecret:  getEnv("POLAR_WEBHOOK_SECRET", ""),
		// Billing system - Polar checkout URLs (no internal product IDs needed)
		PolarPulseSmallURL:  getEnv("POLAR_PULSE_SMALL_URL", ""),
		PolarPulseMediumURL: getEnv("POLAR_PULSE_MEDIUM_URL", ""),
		PolarPulseLargeURL:  getEnv("POLAR_PULSE_LARGE_URL", ""),
		PolarProMonthlyURL:  getEnv("POLAR_PRO_MONTHLY_URL", ""),
		PolarProAnnualURL:   getEnv("POLAR_PRO_ANNUAL_URL", ""),
		PolarEnterpriseURL:  getEnv("POLAR_ENTERPRISE_URL", ""),

		TelegramBotToken:   getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramWebhookURL: getEnv("TELEGRAM_WEBHOOK_URL", ""),

		MetaAccessToken:    getEnv("META_ACCESS_TOKEN", ""),
		MetaPhoneNumberID:  getEnv("META_PHONE_NUMBER_ID", ""),
		MetaPageID:         getEnv("META_PAGE_ID", ""),
		InstagramAccountID: getEnv("INSTAGRAM_ACCOUNT_ID", ""),
		MetaVerifyToken:    getEnv("META_VERIFY_TOKEN", ""),

		OpenWAEnabled:              getEnv("OPENWA_ENABLED", "true") == "true",
		OpenWABaseURL:              getEnv("OPENWA_BASE_URL", "http://localhost:2785"),
		OpenWAApiKey:               getEnv("OPENWA_API_KEY", ""),
		OpenWASessionID:            getEnv("OPENWA_SESSION_ID", "noant-business"),
		OpenWAWebhookSecret:        getEnv("OPENWA_WEBHOOK_SECRET", ""),
		OpenWARateLimitText:        atoiDefault(getEnv("OPENWA_RATE_LIMIT_TEXT", "20")),
		OpenWARateLimitMedia:       atoiDefault(getEnv("OPENWA_RATE_LIMIT_MEDIA", "10")),
		OpenWARateLimitTemplate:    atoiDefault(getEnv("OPENWA_RATE_LIMIT_TEMPLATE", "30")),
		OpenWARateLimitBurst:       atoiDefault(getEnv("OPENWA_RATE_LIMIT_BURST", "5")),
		OpenWAQueueDepth:           atoiDefault(getEnv("OPENWA_QUEUE_DEPTH", "10000")),
		OpenWAMediaDir:             getEnv("OPENWA_MEDIA_DIR", "./media"),
		OpenWAMediaRetention:       time.Duration(atoiDefault(getEnv("OPENWA_MEDIA_RETENTION_DAYS", "90"))) * 24 * time.Hour,
		OpenWASessionHealthInterval: time.Duration(atoiDefault(getEnv("OPENWA_SESSION_HEALTH_INTERVAL", "30"))) * time.Second,
		OpenWAMaxReconnectAttempts: atoiDefault(getEnv("OPENWA_MAX_RECONNECT_ATTEMPTS", "10")),
		OpenWAConnPoolSize:         atoiDefault(getEnv("OPENWA_CONN_POOL_SIZE", "10")),
		OpenWAConnTimeout:          time.Duration(atoiDefault(getEnv("OPENWA_CONN_TIMEOUT", "30"))) * time.Second,
		OpenWAReqTimeout:           time.Duration(atoiDefault(getEnv("OPENWA_REQ_TIMEOUT", "60"))) * time.Second,
	}

	cfg.CORSOrigins = parseCSVEnv("CORS_ORIGINS", cfg.APIURL)

	cfg.Validate()
	return cfg
}

func (c *Config) Validate() {
	if len(strings.TrimSpace(c.JWTSecret)) < 32 {
		fmt.Fprintf(os.Stderr, "FATAL: JWT_SECRET must be set and at least 32 characters long\n")
		os.Exit(1)
	}
	if len(strings.TrimSpace(c.SessionSecret)) < 32 {
		fmt.Fprintf(os.Stderr, "FATAL: SESSION_SECRET must be set and at least 32 characters long\n")
		os.Exit(1)
	}
	if strings.TrimSpace(c.TiDBHost) == "" {
		fmt.Fprintf(os.Stderr, "FATAL: TIDB_HOST environment variable is required\n")
		os.Exit(1)
	}
	if strings.TrimSpace(c.RedisHost) == "" {
		fmt.Fprintf(os.Stderr, "FATAL: REDIS_HOST environment variable is required\n")
		os.Exit(1)
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func atoiDefault(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func parseCSVEnv(key string, fallback string) []string {
	raw := strings.TrimSpace(getEnv(key, ""))
	if raw == "" {
		return []string{fallback}
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	if len(values) == 0 {
		return []string{fallback}
	}
	return values
}
