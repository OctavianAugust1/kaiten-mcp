package kaiten

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"testing"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

func TestWorkspaceReadersUseFixedDocumentedPaths(t *testing.T) {
	t.Parallel()

	var mutex sync.Mutex
	var requests []string
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		requests = append(requests, request.URL.RequestURI())
		mutex.Unlock()
		writer.Header().Set("Content-Type", "application/json")
		if request.URL.Path == "/api/v1/spaces/7" || request.URL.Path == "/api/v1/boards/11" {
			_, _ = writer.Write([]byte(`{"id":1,"extra":"preserved"}`))
			return
		}
		_, _ = writer.Write([]byte(`[{"id":1,"extra":"preserved"}]`))
	}))
	defer server.Close()
	client := newTestClient(t, server)
	page, err := domain.NewPage(20, 5, 50)
	if err != nil {
		t.Fatalf("NewPage() error = %v", err)
	}

	spaces, err := client.ListSpaces(context.Background(), page)
	if err != nil {
		t.Fatalf("ListSpaces() error = %v", err)
	}
	if spaces.Pagination == nil || spaces.Pagination.Limit != 20 || spaces.Pagination.Offset == nil || *spaces.Pagination.Offset != 5 {
		t.Errorf("ListSpaces() pagination = %#v, want limit 20 offset 5", spaces.Pagination)
	}
	spaceID, _ := domain.NewID(7)
	boardID, _ := domain.NewID(11)
	columnID, _ := domain.NewID(13)
	if result, err := client.GetSpace(context.Background(), spaceID); err != nil || result.Data["extra"] != "preserved" {
		t.Fatalf("GetSpace() result/error = %#v/%v", result, err)
	}
	if _, err := client.ListBoards(context.Background(), spaceID); err != nil {
		t.Fatalf("ListBoards() error = %v", err)
	}
	if _, err := client.GetBoard(context.Background(), boardID); err != nil {
		t.Fatalf("GetBoard() error = %v", err)
	}
	if _, err := client.ListColumns(context.Background(), boardID); err != nil {
		t.Fatalf("ListColumns() error = %v", err)
	}
	if _, err := client.ListSubcolumns(context.Background(), columnID); err != nil {
		t.Fatalf("ListSubcolumns() error = %v", err)
	}
	if _, err := client.ListLanes(context.Background(), boardID); err != nil {
		t.Fatalf("ListLanes() error = %v", err)
	}

	sort.Strings(requests)
	want := []string{
		"/api/v1/boards/11",
		"/api/v1/boards/11/columns",
		"/api/v1/boards/11/lanes",
		"/api/v1/columns/13/subcolumns",
		"/api/v1/spaces/7",
		"/api/v1/spaces/7/boards",
		"/api/v1/spaces?limit=20&offset=5",
	}
	sort.Strings(want)
	if len(requests) != len(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	for index := range want {
		if requests[index] != want[index] {
			t.Fatalf("requests = %v, want %v", requests, want)
		}
	}
}
