package app

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/OctavianAugust1/kaiten-mcp/internal/config"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestNewBuildsToolsOnlyServerWithManualDependencies(t *testing.T) {
	t.Setenv("KAITEN_BASE_URL", "https://example.kaiten.ru/api/v1")
	t.Setenv("KAITEN_TOKEN", "test-token")
	t.Setenv("KAITEN_REQUEST_TIMEOUT", "1s")
	t.Setenv("KAITEN_DEFAULT_PAGE_LIMIT", "20")
	t.Setenv("LOG_LEVEL", "error")
	cfg, err := config.Load(filepath.Join(t.TempDir(), "missing.env"))
	if err != nil {
		t.Fatal(err)
	}
	application, err := New(cfg, "v0.1.0-test", io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	serverTransport, clientTransport := sdk.NewInMemoryTransports()
	serverSession, err := application.Server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer serverSession.Close()
	clientSession, err := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "v"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer clientSession.Close()
	tools, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools.Tools) != 23 {
		t.Errorf("tool count = %d, want 23", len(tools.Tools))
	}
}
