package kaiten

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

func TestCardReadersUseFixedPathsAndSinglePage(t *testing.T) {
	t.Parallel()

	var listCalls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/v1/cards":
			listCalls.Add(1)
			if request.URL.Query().Get("version") == "2" {
				_, _ = writer.Write([]byte(`{"result":[{"id":2,"extra":"v2"}],"position":"next-cursor"}`))
				return
			}
			_, _ = writer.Write([]byte(`[{"id":1,"extra":"v1"}]`))
		case "/api/v1/cards/7":
			_, _ = writer.Write([]byte(`{"id":7,"extra":"detail"}`))
		case "/api/v1/cards/7/comments":
			_, _ = writer.Write([]byte(`[{"id":8,"extra":"comment"}]`))
		case "/api/v1/cards/7/members":
			_, _ = writer.Write([]byte(`[{"id":9,"extra":"member"}]`))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()
	client := newTestClient(t, server)

	v1, err := client.ListCards(context.Background(), domain.CardQuery{Limit: 20, Offset: 5}, 50)
	if err != nil {
		t.Fatalf("ListCards(v1) error = %v", err)
	}
	if v1.Data[0]["extra"] != "v1" || v1.Pagination == nil || v1.Pagination.Offset == nil || *v1.Pagination.Offset != 5 {
		t.Errorf("ListCards(v1) = %#v", v1)
	}
	v2, err := client.ListCards(context.Background(), domain.CardQuery{Version: 2, Limit: 20, StartPosition: "current"}, 50)
	if err != nil {
		t.Fatalf("ListCards(v2) error = %v", err)
	}
	if v2.Data[0]["extra"] != "v2" || v2.Pagination == nil || v2.Pagination.NextPosition != "next-cursor" {
		t.Errorf("ListCards(v2) = %#v", v2)
	}
	if listCalls.Load() != 2 {
		t.Errorf("list calls = %d, want exactly one per invocation", listCalls.Load())
	}

	cardID, _ := domain.NewID(7)
	if result, err := client.GetCard(context.Background(), cardID); err != nil || result.Data["extra"] != "detail" {
		t.Errorf("GetCard() result/error = %#v/%v", result, err)
	}
	if result, err := client.ListCardComments(context.Background(), cardID); err != nil || result.Data[0]["extra"] != "comment" {
		t.Errorf("ListCardComments() result/error = %#v/%v", result, err)
	}
	if result, err := client.ListCardMembers(context.Background(), cardID); err != nil || result.Data[0]["extra"] != "member" {
		t.Errorf("ListCardMembers() result/error = %#v/%v", result, err)
	}
}
