package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testBaseURL = "https://example.kaiten.ru/api/v1"
	testToken   = "test-token-that-must-not-leak"
)

func TestLoadRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr string
	}{
		{name: "missing base URL", key: "KAITEN_BASE_URL", value: "", wantErr: "KAITEN_BASE_URL"},
		{name: "missing token", key: "KAITEN_TOKEN", value: "", wantErr: "KAITEN_TOKEN"},
		{name: "HTTP base URL", key: "KAITEN_BASE_URL", value: "http://example.kaiten.ru/api/v1", wantErr: "HTTPS"},
		{name: "relative base URL", key: "KAITEN_BASE_URL", value: "/api/v1", wantErr: "absolute"},
		{name: "base URL user info", key: "KAITEN_BASE_URL", value: "https://user@example.kaiten.ru/api/v1", wantErr: "user info"},
		{name: "base URL query", key: "KAITEN_BASE_URL", value: "https://example.kaiten.ru/api/v1?x=1", wantErr: "query"},
		{name: "base URL fragment", key: "KAITEN_BASE_URL", value: "https://example.kaiten.ru/api/v1#x", wantErr: "fragment"},
		{name: "invalid timeout", key: "KAITEN_REQUEST_TIMEOUT", value: "soon", wantErr: "KAITEN_REQUEST_TIMEOUT"},
		{name: "zero timeout", key: "KAITEN_REQUEST_TIMEOUT", value: "0s", wantErr: "positive"},
		{name: "invalid page limit", key: "KAITEN_DEFAULT_PAGE_LIMIT", value: "many", wantErr: "KAITEN_DEFAULT_PAGE_LIMIT"},
		{name: "zero page limit", key: "KAITEN_DEFAULT_PAGE_LIMIT", value: "0", wantErr: "between 1 and 100"},
		{name: "oversized page limit", key: "KAITEN_DEFAULT_PAGE_LIMIT", value: "101", wantErr: "between 1 and 100"},
		{name: "invalid log level", key: "LOG_LEVEL", value: "trace", wantErr: "LOG_LEVEL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setValidEnvironment(t)
			t.Setenv(tt.key, tt.value)

			_, err := Load(filepath.Join(t.TempDir(), "missing.env"))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Load() error = %v, want error containing %q", err, tt.wantErr)
			}
			if strings.Contains(err.Error(), testToken) {
				t.Fatal("Load() error exposed the token")
			}
		})
	}
}

func TestLoadUsesDefaults(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("KAITEN_REQUEST_TIMEOUT", "")
	t.Setenv("KAITEN_DEFAULT_PAGE_LIMIT", "")
	t.Setenv("LOG_LEVEL", "")

	cfg, err := Load(filepath.Join(t.TempDir(), "missing.env"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got, want := cfg.RequestTimeout(), 30*time.Second; got != want {
		t.Errorf("RequestTimeout() = %s, want %s", got, want)
	}
	if got, want := cfg.DefaultPageLimit(), 50; got != want {
		t.Errorf("DefaultPageLimit() = %d, want %d", got, want)
	}
	if got, want := cfg.LogLevel(), slog.LevelInfo; got != want {
		t.Errorf("LogLevel() = %s, want %s", got, want)
	}
}

func TestLoadReadsDotenvWithoutOverwritingProcessEnvironment(t *testing.T) {
	clearConfigEnvironment(t)

	dotenvPath := filepath.Join(t.TempDir(), ".env")
	contents := strings.Join([]string{
		"KAITEN_BASE_URL=https://dotenv.kaiten.ru/api/latest",
		"KAITEN_TOKEN=dotenv-token",
		"KAITEN_REQUEST_TIMEOUT=45s",
		"KAITEN_DEFAULT_PAGE_LIMIT=25",
		"LOG_LEVEL=debug",
		"",
	}, "\n")
	if err := os.WriteFile(dotenvPath, []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("KAITEN_TOKEN", "process-token")

	cfg, err := Load(dotenvPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	baseURL := cfg.BaseURL()
	if got, want := baseURL.String(), "https://dotenv.kaiten.ru/api/latest/"; got != want {
		t.Errorf("BaseURL() = %q, want %q", got, want)
	}
	if got, want := cfg.Token(), "process-token"; got != want {
		t.Errorf("Token() = %q, want %q", got, want)
	}
	if got, want := cfg.RequestTimeout(), 45*time.Second; got != want {
		t.Errorf("RequestTimeout() = %s, want %s", got, want)
	}
	if got, want := cfg.DefaultPageLimit(), 25; got != want {
		t.Errorf("DefaultPageLimit() = %d, want %d", got, want)
	}
	if got, want := cfg.LogLevel(), slog.LevelDebug; got != want {
		t.Errorf("LogLevel() = %s, want %s", got, want)
	}
}

func setValidEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("KAITEN_BASE_URL", testBaseURL)
	t.Setenv("KAITEN_TOKEN", testToken)
	t.Setenv("KAITEN_REQUEST_TIMEOUT", "10s")
	t.Setenv("KAITEN_DEFAULT_PAGE_LIMIT", "20")
	t.Setenv("LOG_LEVEL", "warn")
}

func clearConfigEnvironment(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"KAITEN_BASE_URL",
		"KAITEN_TOKEN",
		"KAITEN_REQUEST_TIMEOUT",
		"KAITEN_DEFAULT_PAGE_LIMIT",
		"LOG_LEVEL",
	} {
		value, exists := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Unsetenv(%q) error = %v", key, err)
		}
		t.Cleanup(func() {
			if exists {
				_ = os.Setenv(key, value)
				return
			}
			_ = os.Unsetenv(key)
		})
	}
}
