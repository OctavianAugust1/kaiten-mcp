package mcptransport

import (
	"context"
	"testing"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestWorkspaceToolsDiscoverAndCallThroughInMemoryTransport(t *testing.T) {
	service := &workspaceStub{}
	server, err := NewServer("v0.1.0-test", 25, Services{Workspace: service})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	ctx := context.Background()
	serverTransport, clientTransport := sdk.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server Connect() error = %v", err)
	}
	defer serverSession.Close()
	client := sdk.NewClient(&sdk.Implementation{Name: "test-client", Version: "v0.1.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client Connect() error = %v", err)
	}
	defer clientSession.Close()

	listed, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	wantNames := []string{
		"kaiten_get_board",
		"kaiten_get_space",
		"kaiten_list_boards",
		"kaiten_list_columns",
		"kaiten_list_lanes",
		"kaiten_list_spaces",
		"kaiten_list_subcolumns",
	}
	if len(listed.Tools) != len(wantNames) {
		t.Fatalf("tool count = %d, want %d", len(listed.Tools), len(wantNames))
	}
	seen := make(map[string]bool, len(listed.Tools))
	for _, tool := range listed.Tools {
		seen[tool.Name] = true
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint || tool.Annotations.DestructiveHint == nil || *tool.Annotations.DestructiveHint {
			t.Errorf("tool %s annotations = %#v, want read-only/non-destructive", tool.Name, tool.Annotations)
		}
	}
	for _, name := range wantNames {
		if !seen[name] {
			t.Errorf("tool %s was not discovered", name)
		}
	}

	calls := []struct {
		name      string
		arguments map[string]any
	}{
		{name: "kaiten_list_spaces", arguments: map[string]any{"limit": 20, "offset": 5}},
		{name: "kaiten_get_space", arguments: map[string]any{"space_id": 7}},
		{name: "kaiten_list_boards", arguments: map[string]any{"space_id": 7}},
		{name: "kaiten_get_board", arguments: map[string]any{"board_id": 11}},
		{name: "kaiten_list_columns", arguments: map[string]any{"board_id": 11}},
		{name: "kaiten_list_subcolumns", arguments: map[string]any{"column_id": 13}},
		{name: "kaiten_list_lanes", arguments: map[string]any{"board_id": 11}},
	}
	for _, call := range calls {
		result, err := clientSession.CallTool(ctx, &sdk.CallToolParams{Name: call.name, Arguments: call.arguments})
		if err != nil {
			t.Errorf("CallTool(%s) protocol error = %v", call.name, err)
			continue
		}
		if result.IsError || result.StructuredContent == nil {
			t.Errorf("CallTool(%s) result = %#v, want structured success", call.name, result)
		}
	}
	if service.spacePage != (domain.PageRequest{Limit: 20, Offset: 5}) {
		t.Errorf("space page = %#v, want limit 20 offset 5", service.spacePage)
	}

	invalid, err := clientSession.CallTool(ctx, &sdk.CallToolParams{
		Name:      "kaiten_get_space",
		Arguments: map[string]any{"space_id": 0, "unexpected": true},
	})
	if err == nil && !invalid.IsError {
		t.Fatalf("invalid strict-schema call unexpectedly succeeded: %#v", invalid)
	}
}

type workspaceStub struct {
	spacePage domain.PageRequest
}

func (stub *workspaceStub) ListSpaces(_ context.Context, page domain.PageRequest) (domain.Result[[]domain.Object], error) {
	stub.spacePage = page
	return domain.Result[[]domain.Object]{Data: []domain.Object{{"kind": "spaces"}}, Pagination: &domain.Pagination{Limit: page.Limit}}, nil
}

func (stub *workspaceStub) GetSpace(context.Context, domain.ID) (domain.Result[domain.Object], error) {
	return detailStub("space"), nil
}

func (stub *workspaceStub) ListBoards(context.Context, domain.ID) (domain.Result[[]domain.Object], error) {
	return listStub("boards"), nil
}

func (stub *workspaceStub) GetBoard(context.Context, domain.ID) (domain.Result[domain.Object], error) {
	return detailStub("board"), nil
}

func (stub *workspaceStub) ListColumns(context.Context, domain.ID) (domain.Result[[]domain.Object], error) {
	return listStub("columns"), nil
}

func (stub *workspaceStub) ListSubcolumns(context.Context, domain.ID) (domain.Result[[]domain.Object], error) {
	return listStub("subcolumns"), nil
}

func (stub *workspaceStub) ListLanes(context.Context, domain.ID) (domain.Result[[]domain.Object], error) {
	return listStub("lanes"), nil
}

func listStub(kind string) domain.Result[[]domain.Object] {
	return domain.Result[[]domain.Object]{Data: []domain.Object{{"kind": kind}}}
}

func detailStub(kind string) domain.Result[domain.Object] {
	return domain.Result[domain.Object]{Data: domain.Object{"kind": kind}}
}
