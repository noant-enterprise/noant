package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type APIVersionInfo struct {
	SupportedVersions []VersionInfo `json:"supported_versions"`
	DefaultVersion    string        `json:"default_version"`
	LatestVersion     string        `json:"latest_version"`
}

type VersionInfo struct {
	Version    string `json:"version"`
	Status     string `json:"status"`
	SunsetDate string `json:"sunset_date,omitempty"`
}

func DeprecatedAPIVersionMiddleware(oldVersion, newVersion string) gin.HandlerFunc {
	docsURL := fmt.Sprintf("https://docs.noant.example.com/api/%s/migration", newVersion)
	return func(c *gin.Context) {
		c.Header("Deprecation", "true")
		c.Header("Sunset", time.Now().AddDate(0, 0, 90).UTC().Format(time.RFC1123))
		c.Header("Link", fmt.Sprintf("<%s>; rel=\"successor-version\"", docsURL))
		c.Header("X-API-Deprecation-Info", fmt.Sprintf("API version %s is deprecated. Migrate to %s.", oldVersion, newVersion))
		c.Next()
	}
}

func APIVersionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		version := extractVersionFromPath(c.Request.URL.Path)
		if version != "" {
			c.Set("apiVersion", version)
			c.Header("X-API-Version", version)
		}
		c.Next()
	}
}

func extractVersionFromPath(path string) string {
	const prefix = "/api/v"
	idx := strings.Index(path, prefix)
	if idx < 0 {
		return ""
	}
	rest := path[idx+len(prefix):]
	end := strings.IndexAny(rest, "/?")
	if end < 0 {
		return rest
	}
	return rest[:end]
}

func GetAPIVersionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		info := APIVersionInfo{
			SupportedVersions: []VersionInfo{
				{Version: "1", Status: "current"},
			},
			DefaultVersion: "1",
			LatestVersion:  "1",
		}
		c.JSON(http.StatusOK, info)
	}
}

func RequireAPIVersion(allowedVersions ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedVersions))
	for _, v := range allowedVersions {
		allowed[v] = struct{}{}
	}

	return func(c *gin.Context) {
		raw, exists := c.Get("apiVersion")
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "API version not found in request path"})
			c.Abort()
			return
		}

		version, ok := raw.(string)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid API version"})
			c.Abort()
			return
		}

		if _, ok := allowed[version]; !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"error":            "Unsupported API version",
				"requested_version": version,
				"allowed_versions":  allowedVersions,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
