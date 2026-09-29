package mcptransport

import (
	"context"
	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"testing"
)

func TestDocumentToolsValidateUUIDAndCallInMemoryService(t *testing.T) {
	server, err := NewServer("v", 25, Services{Document: &documentStub{}})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	st, ct := sdk.NewInMemoryTransports()
	ss, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	cs, err := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "v"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()
	uid := "550e8400-e29b-41d4-a716-446655440000"
	for _, call := range []struct {
		name string
		args map[string]any
	}{
		{"kaiten_list_documents", map[string]any{"limit": 20, "version": 2, "start_position": "next"}}, {"kaiten_get_document", map[string]any{"document_uid": uid}}, {"kaiten_get_document_schema", map[string]any{"id": 7}}, {"kaiten_list_document_groups", nil}, {"kaiten_get_document_group", map[string]any{"document_group_uid": uid}},
	} {
		result, err := cs.CallTool(ctx, &sdk.CallToolParams{Name: call.name, Arguments: call.args})
		if err != nil || result.IsError {
			t.Errorf("%s: %#v/%v", call.name, result, err)
		}
	}
	invalid, err := cs.CallTool(ctx, &sdk.CallToolParams{Name: "kaiten_get_document", Arguments: map[string]any{"document_uid": "not-a-uuid"}})
	if err == nil && !invalid.IsError {
		t.Fatal("invalid UUID was accepted")
	}
}

type documentStub struct{}

func (*documentStub) ListDocuments(context.Context, domain.DocumentQuery) (domain.Result[[]domain.Object], error) {
	return listStub("documents"), nil
}
func (*documentStub) GetDocument(context.Context, domain.UID) (domain.Result[domain.Object], error) {
	return detailStub("document"), nil
}
func (*documentStub) GetDocumentSchema(context.Context, domain.ID) (domain.Result[domain.Object], error) {
	return detailStub("schema"), nil
}
func (*documentStub) ListDocumentGroups(context.Context) (domain.Result[[]domain.Object], error) {
	return listStub("groups"), nil
}
func (*documentStub) GetDocumentGroup(context.Context, domain.UID) (domain.Result[domain.Object], error) {
	return detailStub("group"), nil
}
