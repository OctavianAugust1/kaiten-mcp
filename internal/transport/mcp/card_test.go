package mcptransport

import (
	"context"
	"testing"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCardToolsDiscoverCallAndRejectBetaFilter(t *testing.T) {
	service := &cardStub{}
	server, err := NewServer("v0.1.0-test", 25, Services{Card: service})
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

	for _, call := range []struct {
		name string
		args map[string]any
	}{
		{"kaiten_list_cards", map[string]any{"query": "urgent", "limit": 20}},
		{"kaiten_get_card", map[string]any{"card_id": 7}},
		{"kaiten_list_card_comments", map[string]any{"card_id": 7}},
		{"kaiten_list_card_members", map[string]any{"card_id": 7}},
	} {
		result, err := clientSession.CallTool(ctx, &sdk.CallToolParams{Name: call.name, Arguments: call.args})
		if err != nil || result.IsError || result.StructuredContent == nil {
			t.Errorf("CallTool(%s) result/error = %#v/%v", call.name, result, err)
		}
	}
	if service.query.Query != "urgent" || service.query.Limit != 20 {
		t.Errorf("list card query = %#v, want query urgent and limit 20", service.query)
	}
	invalid, err := clientSession.CallTool(ctx, &sdk.CallToolParams{
		Name:      "kaiten_list_cards",
		Arguments: map[string]any{"filter": "base64-beta-expression"},
	})
	if err == nil && !invalid.IsError {
		t.Fatalf("beta filter unexpectedly accepted: %#v", invalid)
	}
}

type cardStub struct {
	query domain.CardQuery
}

func (stub *cardStub) ListCards(_ context.Context, query domain.CardQuery) (domain.Result[[]domain.Object], error) {
	stub.query = query
	return listStub("cards"), nil
}

func (*cardStub) GetCard(context.Context, domain.ID) (domain.Result[domain.Object], error) {
	return detailStub("card"), nil
}

func (*cardStub) ListCardComments(context.Context, domain.ID) (domain.Result[[]domain.Object], error) {
	return listStub("comments"), nil
}

func (*cardStub) ListCardMembers(context.Context, domain.ID) (domain.Result[[]domain.Object], error) {
	return listStub("members"), nil
}
