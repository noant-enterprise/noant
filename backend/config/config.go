package config

import (
	"github.com/joho/godotenv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port            string
	JWTSecret       string
	SessionSecret   string
	NodeEnv         string
	LogLevel        string
	APIURL          string

	// Cache
	CacheTTL      time.Duration
	CacheMaxKeys  int

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
	TwilioAccountSID    string
	TwilioAuthToken     string
	TwilioWhatsAppNumber string

	// Resend
	ResendAPIKey string

	// Polar
	PolarAccessToken     string
	PolarOrganizationID  string
	PolarServerURL       string
	PolarWebhookSecret   string
	CORSOrigins         []string

	// Telegram
	TelegramBotToken string
	TelegramWebhookURL string

	// Meta (WhatsApp / Facebook / Instagram)
	MetaAccessToken    string
	MetaPhoneNumberID  string
	MetaPageID         string
	InstagramAccountID string
	MetaVerifyToken    string
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

	tidbPort, _ := strconv.Atoi(getEnv("TIDB_PORT", "4000"))
	dbPoolSize, _ := strconv.Atoi(getEnv("DB_POOL_SIZE", "20"))

	redisPort, _ := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	redisSSL, _ := strconv.ParseBool(getEnv("REDIS_SSL", "true"))
	redisShortTTL, _ := strconv.Atoi(getEnv("REDIS_SHORT_TTL", "259200"))

	var groqKeys []string
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

		PolarAccessToken:    getEnv("POLAR_ACCESS_TOKEN", ""),
		PolarOrganizationID: getEnv("POLAR_ORGANIZATION_ID", ""),
		PolarServerURL:      getEnv("POLAR_SERVER_URL", "https://api.polar.sh"),
		PolarWebhookSecret:  getEnv("POLAR_WEBHOOK_SECRET", ""),

		TelegramBotToken:   getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramWebhookURL: getEnv("TELEGRAM_WEBHOOK_URL", ""),

		MetaAccessToken:    getEnv("META_ACCESS_TOKEN", ""),
		MetaPhoneNumberID:  getEnv("META_PHONE_NUMBER_ID", ""),
		MetaPageID:         getEnv("META_PAGE_ID", ""),
		InstagramAccountID: getEnv("INSTAGRAM_ACCOUNT_ID", ""),
		MetaVerifyToken:    getEnv("META_VERIFY_TOKEN", ""),
	}

	// Parse CORS origins
	if origins := getEnv("CORS_ORIGINS", ""); origins != "" {
		cfg.CORSOrigins = strings.Split(origins, ",")
	} else {
		cfg.CORSOrigins = []string{cfg.APIURL}
	}

	// Parse CORS origins
	if origins := getEnv("CORS_ORIGINS", ""); origins != "" {
		cfg.CORSOrigins = strings.Split(origins, ",")
	} else {
		cfg.CORSOrigins = []string{cfg.APIURL}
	}

	// Parse CORS origins
	if origins := getEnv("CORS_ORIGINS", ""); origins != "" {
		cfg.CORSOrigins = strings.Split(origins, ",")
	} else {
		cfg.CORSOrigins = []string{cfg.APIURL}
	}

	cfg.Validate()
	return cfg
}

func (c *Config) Validate() {
	if c.JWTSecret == "" {
		panic("FATAL: JWT_SECRET environment variable is required")
	}
	if c.SessionSecret == "" {
		panic("FATAL: SESSION_SECRET environment variable is required")
	}
	if c.TiDBHost == "" {
		panic("FATAL: TIDB_HOST environment variable is required")
	}
	if c.RedisHost == "" {
		panic("FATAL: REDIS_HOST environment variable is required")
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
