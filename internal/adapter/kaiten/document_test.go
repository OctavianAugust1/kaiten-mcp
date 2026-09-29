package kaiten

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

func TestDocumentReadersPreservePayloadAndContinuation(t *testing.T) {
	t.Parallel()
	var documentCalls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/v1/documents":
			documentCalls.Add(1)
			if request.URL.Query().Get("version") == "2" {
				_, _ = writer.Write([]byte(`{"result":[{"uid":"x","nested":{"kept":true}}],"position":"next"}`))
				return
			}
			_, _ = writer.Write([]byte(`[{"uid":"x","nested":{"kept":true}}]`))
		case "/api/v1/documents/550e8400-e29b-41d4-a716-446655440000", "/api/v1/document-schemas/7", "/api/v1/document-groups/550e8400-e29b-41d4-a716-446655440000":
			_, _ = writer.Write([]byte(`{"nested":{"kept":true}}`))
		case "/api/v1/document-groups":
			_, _ = writer.Write([]byte(`[{"nested":{"kept":true}}]`))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()
	client := newTestClient(t, server)
	v1, err := client.ListDocuments(context.Background(), domain.DocumentQuery{Limit: 20, Offset: 5}, 50)
	if err != nil {
		t.Fatal(err)
	}
	v2, err := client.ListDocuments(context.Background(), domain.DocumentQuery{Version: 2, Limit: 20, StartPosition: "current"}, 50)
	if err != nil {
		t.Fatal(err)
	}
	if v1.Data[0]["nested"].(map[string]any)["kept"] != true || v2.Pagination.NextPosition != "next" || documentCalls.Load() != 2 {
		t.Errorf("list outputs/calls = %#v/%#v/%d", v1, v2, documentCalls.Load())
	}
	uid, _ := domain.ParseUID("550e8400-e29b-41d4-a716-446655440000")
	schemaID, _ := domain.NewID(7)
	if _, err := client.GetDocument(context.Background(), uid); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetDocumentSchema(context.Background(), schemaID); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListDocumentGroups(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetDocumentGroup(context.Background(), uid); err != nil {
		t.Fatal(err)
	}
}
