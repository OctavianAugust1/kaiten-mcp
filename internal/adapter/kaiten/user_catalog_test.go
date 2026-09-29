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

func TestUserAndCatalogReadersUseFixedGETPaths(t *testing.T) {
	t.Parallel()
	var mutex sync.Mutex
	var requests []string
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		requests = append(requests, request.URL.RequestURI())
		mutex.Unlock()
		if request.URL.Path == "/api/v1/users/current" || request.URL.Path == "/api/v1/spaces/7/users/8" || request.URL.Path == "/api/v1/card-types/9" {
			_, _ = writer.Write([]byte(`{"id":1,"extra":"detail"}`))
			return
		}
		_, _ = writer.Write([]byte(`[{"id":1,"extra":"list"}]`))
	}))
	defer server.Close()
	client := newTestClient(t, server)
	page, _ := domain.NewPage(20, 5, 50)
	if _, err := client.ListUsers(context.Background(), domain.UserQuery{Page: page, Query: "alex", IncludeInactive: boolPointer(true)}); err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	spaceID, _ := domain.NewID(7)
	userID, _ := domain.NewID(8)
	typeID, _ := domain.NewID(9)
	if _, err := client.GetCurrentUser(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListSpaceUsers(context.Background(), spaceID); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetSpaceUser(context.Background(), spaceID, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListCardTypes(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetCardType(context.Background(), typeID); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListTags(context.Background()); err != nil {
		t.Fatal(err)
	}
	sort.Strings(requests)
	want := []string{
		"/api/v1/card-types", "/api/v1/card-types/9", "/api/v1/spaces/7/users", "/api/v1/spaces/7/users/8", "/api/v1/tags", "/api/v1/users/current", "/api/v1/users?include_inactive=true&limit=20&offset=5&query=alex",
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

func boolPointer(value bool) *bool { return &value }
