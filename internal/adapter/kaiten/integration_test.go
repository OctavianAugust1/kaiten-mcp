//go:build integration

package kaiten

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

// TestIntegrationListSpaces exercises an authenticated production-like GET
// only when explicitly enabled with -tags=integration and credentials are set.
func TestIntegrationListSpaces(t *testing.T) {
	baseURL := strings.TrimSpace(os.Getenv("KAITEN_BASE_URL"))
	token := strings.TrimSpace(os.Getenv("KAITEN_TOKEN"))
	if baseURL == "" || token == "" {
		t.Skip("KAITEN_BASE_URL and KAITEN_TOKEN are required for integration testing")
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("parse KAITEN_BASE_URL: %v", err)
	}
	client, err := NewClient(
		*parsedURL,
		token,
		"integration-test",
		&http.Client{Transport: NewReadOnlyTransport(http.DefaultTransport)},
		30*time.Second,
	)
	if err != nil {
		t.Fatalf("create Kaiten client: %v", err)
	}

	page, err := domain.NewPage(1, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListSpaces(context.Background(), page); err != nil {
		t.Fatalf("list spaces through GET-only adapter: %v", err)
	}
}
