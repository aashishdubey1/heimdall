package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Database DBConfig
	Logging  LoggingConfig
}
type AppConfig struct {
	Name        string
	Environment string
	Version     string
}

type HTTPConfig struct {
	Addr              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxHeaderBytes    int
	MaxBodyBytes      int
}

type DBConfig struct {
	URL               string
	MaxConns          int
	MinConns          int
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
}

type LoggingConfig struct {
	Level  string
	Format string
}

func Load() (Config, error) {

	var cfg Config
	var err error

	////////////// -----------> App
	cfg.App = AppConfig{
		Name:        getEnv("APP_NAME", "Heimdall"),
		Environment: getEnv("APP_ENV", "development"),
		Version:     getEnv("APP_VERSION", "dev"),
	}

	////////////// -----------> HTTP
	cfg.HTTP = HTTPConfig{
		Addr:           getEnv("HTTP_ADDR", ":8080"),
		MaxHeaderBytes: 1 << 20,
		MaxBodyBytes:   10 << 20,
	}

	if cfg.HTTP.ReadHeaderTimeout, err = durationEnv(
		"HTTP_READ_HEADER_TIMEOUT", 5*time.Second,
	); err != nil {
		return Config{}, err
	}

	if cfg.HTTP.ReadTimeout, err = durationEnv(
		"HTTP_READ_TIMEOUT", 15*time.Second,
	); err != nil {
		return Config{}, err
	}

	if cfg.HTTP.WriteTimeout, err = durationEnv(
		"HTTP_WRITE_TIMEOUT", 30*time.Second,
	); err != nil {
		return Config{}, err
	}

	if cfg.HTTP.IdleTimeout, err = durationEnv(
		"HTTP_IDLE_TIMEOUT", 60*time.Second,
	); err != nil {
		return Config{}, err
	}

	if cfg.HTTP.ShutdownTimeout, err = durationEnv(
		"HTTP_SHUTDOWN_TIMEOUT", 10*time.Second,
	); err != nil {
		return Config{}, err
	}

	if cfg.HTTP.MaxHeaderBytes, err = intEnv(
		"HTTP_MAX_HEADER_BYTES", 1<<20,
	); err != nil {
		return Config{}, err
	}

	if cfg.HTTP.MaxBodyBytes, err = intEnv(
		"HTTP_MAX_BODY_BYTES", 10<<20,
	); err != nil {
		return Config{}, err
	}

	/////////--------> DATABASESSS
	cfg.Database.URL = strings.TrimSpace(
		os.Getenv("DATABASE_URL"),
	)

	if cfg.Database.MaxConns, err = intEnv(
		"DB_MAX_CONNS", 20,
	); err != nil {
		return Config{}, err
	}

	if cfg.Database.MinConns, err = intEnv(
		"DB_MIN_CONNS", 2,
	); err != nil {
		return Config{}, err
	}

	if cfg.Database.MaxConnLifetime, err = durationEnv(
		"DB_MAX_CONN_LIFETIME", time.Hour,
	); err != nil {
		return Config{}, err
	}

	if cfg.Database.MaxConnIdleTime, err = durationEnv(
		"DB_MAX_CONN_IDLE_TIME", 30*time.Minute,
	); err != nil {
		return Config{}, err
	}

	if cfg.Database.HealthCheckPeriod, err = durationEnv(
		"DB_HEALTH_CHECK_PERIOD", time.Minute,
	); err != nil {
		return Config{}, err
	}

	if cfg.Database.ConnectTimeout, err = durationEnv(
		"DB_CONNECT_TIMEOUT", 5*time.Second,
	); err != nil {
		return Config{}, err
	}

	//////////// --------------> LOGGING
	cfg.Logging = LoggingConfig{
		Level:  strings.ToLower(getEnv("LOG_LEVEL", "info")),
		Format: strings.ToLower(getEnv("LOG_FORMAT", "json")),
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("APP_NAME cannot be empty")
	}

	switch c.App.Environment {
	case "development", "test", "staging", "production":
	default:
		return fmt.Errorf(
			"APP_ENV must be development, test, staging, or production",
		)
	}

	if c.HTTP.Addr == "" {
		return fmt.Errorf("HTTP_ADDR cannot be empty")
	}

	if c.HTTP.ReadHeaderTimeout <= 0 ||
		c.HTTP.ReadTimeout <= 0 ||
		c.HTTP.WriteTimeout <= 0 ||
		c.HTTP.IdleTimeout <= 0 ||
		c.HTTP.ShutdownTimeout <= 0 {
		return fmt.Errorf("HTTP timeouts must be positive")
	}

	if c.HTTP.MaxHeaderBytes <= 0 || c.HTTP.MaxBodyBytes <= 0 {
		return fmt.Errorf("HTTP size limits must be positive")
	}

	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	dbURL, err := url.Parse(c.Database.URL)
	if err != nil ||
		(dbURL.Scheme != "postgres" &&
			dbURL.Scheme != "postgresql") ||
		dbURL.Host == "" {
		return fmt.Errorf("DATABASE_URL must be a valid PostgreSQL URL")
	}

	if c.Database.MinConns < 0 ||
		c.Database.MaxConns <= 0 ||
		c.Database.MinConns > c.Database.MaxConns {
		return fmt.Errorf("invalid database connection pool limits")
	}

	if c.Database.MaxConnLifetime <= 0 ||
		c.Database.MaxConnIdleTime <= 0 ||
		c.Database.HealthCheckPeriod <= 0 ||
		c.Database.ConnectTimeout <= 0 {
		return fmt.Errorf("database durations must be positive")
	}

	switch c.Logging.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("LOG_LEVEL must be debug, info, warn, or error")
	}

	switch c.Logging.Format {
	case "json", "text":
	default:
		return fmt.Errorf("LOG_FORMAT must be json or text")
	}
	return nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid duration: %w", key, err)
	}
	return d, nil
}

func intEnv(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid integer: %w", key, err)
	}
	return n, nil
}
