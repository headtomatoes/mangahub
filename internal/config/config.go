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
	// Environment
	GoEnv   string `env:"GO_ENV" default:"development"`
	NodeEnv string `env:"NODE_ENV" default:"development"`

	// Service Ports
	HTTPPort int `env:"HTTP_PORT" default:"8080"`
	TCPPort  int `env:"TCP_PORT" default:"8081"`
	UDPPort  int `env:"UDP_PORT" default:"8082"`
	GRPCPort int `env:"GRPC_PORT" default:"8083"`

	// Database
	DatabaseURL string `env:"DATABASE_URL" default:"/app/data/mangahub.db"`
	SQLitePath  string `env:"SQLITE_PATH" default:"/app/data/mangahub.db"` //(redundant now)

	// Authentication
	JWTSecret string        `env:"JWT_SECRET" required:"true"`
	JWTExpiry time.Duration `env:"JWT_EXPIRY" default:"24h"`

	// Token TTLs
	AccessTokenTTL  time.Duration `env:"ACCESS_TOKEN_TTL" required:"true" default:"15m"`
	RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL" required:"true" default:"7day"`

	// Redis Cache
	RedisURL      string `env:"REDIS_URL" default:"redis://redis:6379"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	CacheTTL      int    `env:"CACHE_TTL" default:"3600"`

	// External APIs
	MangaDXAPIURL string `env:"MANGADX_API_URL" default:"https://api.mangadx.org"`
	MangaDXAPIKey string `env:"MANGADX_API_KEY"`

	// Monitoring
	PrometheusEnabled bool   `env:"PROMETHEUS_ENABLED" default:"false"`
	GrafanaPassword   string `env:"GRAFANA_PASSWORD" default:"admin"`

	// Development
	LogLevel    string   `env:"LOG_LEVEL" default:"debug"`
	LogFormat   string   `env:"LOG_FORMAT" default:"text"`
	CORSOrigins []string `env:"CORS_ORIGINS" default:"http://localhost:3000,http://localhost:8084"`

	// File Storage
	MangaDataPath string `env:"MANGA_DATA_PATH" default:"/app/data/manga"`
	UserDataPath  string `env:"USER_DATA_PATH" default:"/app/data/users"`
	UploadMaxSize string `env:"UPLOAD_MAX_SIZE" default:"10MB"`

	// TLS
	TLSEnabled  bool   `env:"TLS_ENABLED" default:"false"`
	TLSCertPath string `env:"TLS_CERT_PATH" default:"./cert/localhost+2.pem"`
	TLSKeyPath  string `env:"TLS_KEY_PATH" default:"./cert/localhost+2-key.pem"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Try to load .env file from project root (adjust path as needed)
	err := godotenv.Load(".env") // Go up two levels from test/config to project root
	if err != nil {
		// If .env file doesn't exist, that's OK - we can still use system env vars
		// Only log this in development, don't fail
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	config := &Config{}

	// Load each field with proper type conversion and validation
	if err := loadEnvString(&config.GoEnv, "GO_ENV", "development"); err != nil {
		return nil, err
	}
	if err := loadEnvString(&config.NodeEnv, "NODE_ENV", "development"); err != nil {
		return nil, err
	}

	// Ports
	if err := loadEnvInt(&config.HTTPPort, "HTTP_PORT", 8080); err != nil {
		return nil, err
	}
	if err := loadEnvInt(&config.TCPPort, "TCP_PORT", 8081); err != nil {
		return nil, err
	}
	if err := loadEnvInt(&config.UDPPort, "UDP_PORT", 8082); err != nil {
		return nil, err
	}
	if err := loadEnvInt(&config.GRPCPort, "GRPC_PORT", 8083); err != nil {
		return nil, err
	}

	// Database
	if err := loadEnvString(&config.DatabaseURL, "DATABASE_URL", "/app/data/mangahub.db"); err != nil {
		return nil, err
	}
	if err := loadEnvString(&config.SQLitePath, "SQLITE_PATH", "/app/data/mangahub.db"); err != nil {
		return nil, err
	}

	// Authentication
	if err := loadEnvStringRequired(&config.JWTSecret, "JWT_SECRET"); err != nil {
		return nil, err
	}
	if err := loadEnvDuration(&config.JWTExpiry, "JWT_EXPIRY", 24*time.Hour); err != nil {
		return nil, err
	}

	// Token TTLs
	if err := loadEnvDuration(&config.AccessTokenTTL, "ACCESS_TOKEN_TTL", 15*time.Minute); err != nil {
		return nil, err
	}

	if err := loadEnvDuration(&config.RefreshTokenTTL, "REFRESH_TOKEN_TTL", 7*24*time.Hour); err != nil {
		return nil, err
	}

	// Redis
	if err := loadEnvString(&config.RedisURL, "REDIS_URL", "redis://redis:6379"); err != nil {
		return nil, err
	}
	if err := loadEnvString(&config.RedisPassword, "REDIS_PASSWORD", ""); err != nil {
		return nil, err
	}
	if err := loadEnvInt(&config.CacheTTL, "CACHE_TTL", 3600); err != nil {
		return nil, err
	}

	// External APIs
	if err := loadEnvString(&config.MangaDXAPIURL, "MANGADX_API_URL", "https://api.mangadx.org"); err != nil {
		return nil, err
	}
	if err := loadEnvString(&config.MangaDXAPIKey, "MANGADX_API_KEY", ""); err != nil {
		return nil, err
	}

	// Monitoring
	if err := loadEnvBool(&config.PrometheusEnabled, "PROMETHEUS_ENABLED", false); err != nil {
		return nil, err
	}
	if err := loadEnvString(&config.GrafanaPassword, "GRAFANA_PASSWORD", "admin"); err != nil {
		return nil, err
	}

	// Development
	if err := loadEnvString(&config.LogLevel, "LOG_LEVEL", "debug"); err != nil {
		return nil, err
	}
	if err := loadEnvString(&config.LogFormat, "LOG_FORMAT", "text"); err != nil {
		return nil, err
	}
	if err := loadEnvStringSlice(&config.CORSOrigins, "CORS_ORIGINS", []string{"http://localhost:3000", "http://localhost:8080"}); err != nil {
		return nil, err
	}

	// File Storage
	if err := loadEnvString(&config.MangaDataPath, "MANGA_DATA_PATH", "/app/data/manga"); err != nil {
		return nil, err
	}
	if err := loadEnvString(&config.UserDataPath, "USER_DATA_PATH", "/app/data/users"); err != nil {
		return nil, err
	}
	if err := loadEnvString(&config.UploadMaxSize, "UPLOAD_MAX_SIZE", "10MB"); err != nil {
		return nil, err
	}

	// TLS
	if err := loadEnvBool(&config.TLSEnabled, "TLS_ENABLED", false); err != nil {
		return nil, err
	}
	if err := loadEnvString(&config.TLSCertPath, "TLS_CERT_PATH", "./cert/localhost+2.pem"); err != nil {
		return nil, err
	}
	if err := loadEnvString(&config.TLSKeyPath, "TLS_KEY_PATH", "./cert/localhost+2-key.pem"); err != nil {
		return nil, err
	}
	return config, nil
}

// Helper functions for type conversion and validation
func loadEnvString(target *string, key, defaultValue string) error {
	if value := os.Getenv(key); value != "" {
		*target = value
	} else {
		*target = defaultValue
	}
	return nil
}

func loadEnvStringRequired(target *string, key string) error {
	value := os.Getenv(key)
	if value == "" {
		return fmt.Errorf("required environment variable %s is not set", key)
	}
	*target = value
	return nil
}

func loadEnvInt(target *int, key string, defaultValue int) error {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value for %s: %v", key, err)
		}
		*target = parsed
	} else {
		*target = defaultValue
	}
	return nil
}

func loadEnvBool(target *bool, key string, defaultValue bool) error {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for %s: %v", key, err)
		}
		*target = parsed
	} else {
		*target = defaultValue
	}
	return nil
}

func loadEnvDuration(target *time.Duration, key string, defaultValue time.Duration) error {
	if value := os.Getenv(key); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid duration value for %s: %v", key, err)
		}
		*target = parsed
	} else {
		*target = defaultValue
	}
	return nil
}

func loadEnvStringSlice(target *[]string, key string, defaultValue []string) error {
	if value := os.Getenv(key); value != "" {
		*target = strings.Split(value, ",")
		// Trim whitespace from each element
		for i, v := range *target {
			(*target)[i] = strings.TrimSpace(v)
		}
	} else {
		*target = defaultValue
	}
	return nil
}

// Validate performs validation on the loaded configuration
func (c *Config) Validate() error {
	var errors []string

	// Validate ports are in valid range
	if c.HTTPPort < 1 || c.HTTPPort > 65535 {
		errors = append(errors, "HTTP_PORT must be between 1 and 65535")
	}
	if c.TCPPort < 1 || c.TCPPort > 65535 {
		errors = append(errors, "TCP_PORT must be between 1 and 65535")
	}
	if c.UDPPort < 1 || c.UDPPort > 65535 {
		errors = append(errors, "UDP_PORT must be between 1 and 65535")
	}
	if c.GRPCPort < 1 || c.GRPCPort > 65535 {
		errors = append(errors, "GRPC_PORT must be between 1 and 65535")
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	if !contains(validLogLevels, c.LogLevel) {
		errors = append(errors, fmt.Sprintf("LOG_LEVEL must be one of: %s", strings.Join(validLogLevels, ", ")))
	}

	// Validate log format
	validLogFormats := []string{"text", "json"}
	if !contains(validLogFormats, c.LogFormat) {
		errors = append(errors, fmt.Sprintf("LOG_FORMAT must be one of: %s", strings.Join(validLogFormats, ", ")))
	}

	// Validate JWT secret length (should be at least 32 characters for security)
	if len(c.JWTSecret) < 32 {
		errors = append(errors, "JWT_SECRET should be at least 32 characters long")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// IsDevelopment returns true if the application is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.GoEnv == "development"
}

// IsProduction returns true if the application is running in production mode
func (c *Config) IsProduction() bool {
	return c.GoEnv == "production"
}

// GetDatabasePath returns the appropriate database path(redundant now)
func (c *Config) GetDatabasePath() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	return c.SQLitePath
}

// Helper function to check if slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
