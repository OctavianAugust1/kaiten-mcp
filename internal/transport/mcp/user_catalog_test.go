package mcptransport

import (
	"context"
	"testing"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestUserAndCatalogToolsCallInMemoryServices(t *testing.T) {
	server, err := NewServer("v0.1.0-test", 25, Services{User: &userStub{}, Catalog: &catalogStub{}})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	serverTransport, clientTransport := sdk.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer serverSession.Close()
	clientSession, err := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "v"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer clientSession.Close()
	for _, call := range []struct {
		name string
		args map[string]any
	}{
		{"kaiten_list_users", map[string]any{"limit": 20, "offset": 0, "query": "alex"}}, {"kaiten_get_current_user", nil},
		{"kaiten_list_space_users", map[string]any{"space_id": 7}}, {"kaiten_get_space_user", map[string]any{"space_id": 7, "user_id": 8}},
		{"kaiten_list_card_types", nil}, {"kaiten_get_card_type", map[string]any{"type_id": 9}}, {"kaiten_list_tags", nil},
	} {
		result, err := clientSession.CallTool(ctx, &sdk.CallToolParams{Name: call.name, Arguments: call.args})
		if err != nil || result.IsError || result.StructuredContent == nil {
			t.Errorf("%s: result/error = %#v/%v", call.name, result, err)
		}
	}
}

func TestListSpaceUsersRejectsParametersOutsideItsSchema(t *testing.T) {
	server, err := NewServer("v0.1.0-test", 25, Services{User: &userStub{}})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	serverTransport, clientTransport := sdk.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer serverSession.Close()
	clientSession, err := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "v"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer clientSession.Close()

	result, err := clientSession.CallTool(ctx, &sdk.CallToolParams{
		Name:      "kaiten_list_space_users",
		Arguments: map[string]any{"space_id": 7, "user_id": 8},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatalf("result.IsError = false, want true; result = %#v", result)
	}
}

type userStub struct{}

func (*userStub) ListUsers(context.Context, domain.UserQuery) (domain.Result[[]domain.Object], error) {
	return listStub("users"), nil
}
func (*userStub) GetCurrentUser(context.Context) (domain.Result[domain.Object], error) {
	return detailStub("current_user"), nil
}
func (*userStub) ListSpaceUsers(context.Context, domain.ID) (domain.Result[[]domain.Object], error) {
	return listStub("space_users"), nil
}
func (*userStub) GetSpaceUser(context.Context, domain.ID, domain.ID) (domain.Result[domain.Object], error) {
	return detailStub("space_user"), nil
}

type catalogStub struct{}

func (*catalogStub) ListCardTypes(context.Context) (domain.Result[[]domain.Object], error) {
	return listStub("card_types"), nil
}
func (*catalogStub) GetCardType(context.Context, domain.ID) (domain.Result[domain.Object], error) {
	return detailStub("card_type"), nil
}
func (*catalogStub) ListTags(context.Context) (domain.Result[[]domain.Object], error) {
	return listStub("tags"), nil
}
