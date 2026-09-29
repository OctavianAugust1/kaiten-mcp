package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultRequestTimeout = 30 * time.Second
	defaultPageLimit      = 50
	maxPageLimit          = 100
)

// Config contains validated process configuration. Its accessors return
// values, so callers cannot mutate the configuration after startup.
type Config struct {
	baseURL          url.URL
	token            string
	requestTimeout   time.Duration
	defaultPageLimit int
	logLevel         slog.Level
}

// Load reads an optional dotenv file, overlays process environment values,
// and validates the complete startup configuration.
func Load(dotenvPath string) (Config, error) {
	dotenv, err := godotenv.Read(dotenvPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("read dotenv file: %w", err)
	}
	if dotenv == nil {
		dotenv = make(map[string]string)
	}

	value := func(key string) string {
		if processValue, exists := os.LookupEnv(key); exists {
			return processValue
		}
		return dotenv[key]
	}

	baseURL, err := parseBaseURL(value("KAITEN_BASE_URL"))
	if err != nil {
		return Config{}, err
	}
	token := strings.TrimSpace(value("KAITEN_TOKEN"))
	if token == "" {
		return Config{}, errors.New("KAITEN_TOKEN is required")
	}
	requestTimeout, err := parseDuration(value("KAITEN_REQUEST_TIMEOUT"))
	if err != nil {
		return Config{}, err
	}
	pageLimit, err := parsePageLimit(value("KAITEN_DEFAULT_PAGE_LIMIT"))
	if err != nil {
		return Config{}, err
	}
	logLevel, err := parseLogLevel(value("LOG_LEVEL"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		baseURL:          baseURL,
		token:            token,
		requestTimeout:   requestTimeout,
		defaultPageLimit: pageLimit,
		logLevel:         logLevel,
	}, nil
}

// BaseURL returns a copy of the normalized Kaiten API base URL.
func (c Config) BaseURL() url.URL {
	return c.baseURL
}

// Token returns the bearer token used only by the Kaiten adapter.
func (c Config) Token() string {
	return c.token
}

// RequestTimeout returns the maximum duration of one upstream attempt.
func (c Config) RequestTimeout() time.Duration {
	return c.requestTimeout
}

// DefaultPageLimit returns the bounded page size used when a tool omits it.
func (c Config) DefaultPageLimit() int {
	return c.defaultPageLimit
}

// LogLevel returns the configured structured logging threshold.
func (c Config) LogLevel() slog.Level {
	return c.logLevel
}

func parseBaseURL(raw string) (url.URL, error) {
	if strings.TrimSpace(raw) == "" {
		return url.URL{}, errors.New("KAITEN_BASE_URL is required")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return url.URL{}, errors.New("KAITEN_BASE_URL must be a valid absolute URL")
	}
	if !parsed.IsAbs() || parsed.Host == "" {
		return url.URL{}, errors.New("KAITEN_BASE_URL must be an absolute URL")
	}
	if parsed.Scheme != "https" {
		return url.URL{}, errors.New("KAITEN_BASE_URL must use HTTPS")
	}
	if parsed.User != nil {
		return url.URL{}, errors.New("KAITEN_BASE_URL must not contain user info")
	}
	if parsed.RawQuery != "" {
		return url.URL{}, errors.New("KAITEN_BASE_URL must not contain a query")
	}
	if parsed.Fragment != "" {
		return url.URL{}, errors.New("KAITEN_BASE_URL must not contain a fragment")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/"
	return *parsed, nil
}

func parseDuration(raw string) (time.Duration, error) {
	if raw == "" {
		return defaultRequestTimeout, nil
	}
	duration, err := time.ParseDuration(raw)
	if err != nil {
		return 0, errors.New("KAITEN_REQUEST_TIMEOUT must be a valid duration")
	}
	if duration <= 0 {
		return 0, errors.New("KAITEN_REQUEST_TIMEOUT must be positive")
	}
	return duration, nil
}

func parsePageLimit(raw string) (int, error) {
	if raw == "" {
		return defaultPageLimit, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, errors.New("KAITEN_DEFAULT_PAGE_LIMIT must be an integer")
	}
	if limit < 1 || limit > maxPageLimit {
		return 0, errors.New("KAITEN_DEFAULT_PAGE_LIMIT must be between 1 and 100")
	}
	return limit, nil
}

func parseLogLevel(raw string) (slog.Level, error) {
	switch strings.ToLower(raw) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, errors.New("LOG_LEVEL must be debug, info, warn, or error")
	}
}
