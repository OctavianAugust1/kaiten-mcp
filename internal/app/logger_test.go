package app

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestNewLoggerWritesStructuredRedactedRecordsOnlyToStderr(t *testing.T) {
	var stderr bytes.Buffer
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error = %v", err)
	}
	originalStdout := os.Stdout
	os.Stdout = stdoutWriter
	t.Cleanup(func() { os.Stdout = originalStdout })

	logger := NewLogger(slog.LevelDebug, &stderr)
	logger.LogAttrs(
		context.Background(),
		slog.LevelInfo,
		"request finished",
		slog.String("token", "top-secret-token"),
		slog.String("Authorization", "Bearer top-secret-token"),
		slog.Group("request", slog.String("client_secret", "top-secret-token")),
		slog.Int("status", 200),
	)

	got := stderr.String()
	if !strings.Contains(got, `"msg":"request finished"`) || !strings.Contains(got, `"status":200`) {
		t.Fatalf("stderr record = %q, want structured message and status", got)
	}
	if strings.Contains(got, "top-secret-token") || strings.Contains(got, "Bearer") {
		t.Fatalf("stderr record exposed a credential: %q", got)
	}
	if count := strings.Count(got, `"[REDACTED]"`); count != 3 {
		t.Errorf("stderr redaction count = %d, want 3; record = %q", count, got)
	}
	if err := stdoutWriter.Close(); err != nil {
		t.Fatalf("stdout Close() error = %v", err)
	}
	var stdout bytes.Buffer
	if _, err := stdout.ReadFrom(stdoutReader); err != nil {
		t.Fatalf("stdout ReadFrom() error = %v", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty protocol stream", stdout.String())
	}
}

func TestNewLoggerHonorsConfiguredLevel(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	logger := NewLogger(slog.LevelWarn, &stderr)
	logger.Info("not emitted")
	logger.Warn("emitted")

	if strings.Contains(stderr.String(), "not emitted") || !strings.Contains(stderr.String(), "emitted") {
		t.Errorf("stderr = %q, want only warning record", stderr.String())
	}
}
