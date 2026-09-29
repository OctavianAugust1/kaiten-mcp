package mcptransport_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/OctavianAugust1/kaiten-mcp/internal/app"
	"github.com/OctavianAugust1/kaiten-mcp/internal/config"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var approvedToolNames = []string{
	"kaiten_get_board", "kaiten_get_card", "kaiten_get_card_type", "kaiten_get_current_user", "kaiten_get_document", "kaiten_get_document_group", "kaiten_get_document_schema", "kaiten_get_space", "kaiten_get_space_user",
	"kaiten_list_boards", "kaiten_list_card_comments", "kaiten_list_card_members", "kaiten_list_card_types", "kaiten_list_cards", "kaiten_list_columns", "kaiten_list_document_groups", "kaiten_list_documents", "kaiten_list_lanes", "kaiten_list_spaces", "kaiten_list_space_users", "kaiten_list_subcolumns", "kaiten_list_tags", "kaiten_list_users",
}

func TestRegistryPolicyAllowsExactlyApprovedReadOnlyTools(t *testing.T) {
	t.Setenv("KAITEN_BASE_URL", "https://example.kaiten.ru/api/v1")
	t.Setenv("KAITEN_TOKEN", "test-token")
	t.Setenv("KAITEN_REQUEST_TIMEOUT", "1s")
	t.Setenv("KAITEN_DEFAULT_PAGE_LIMIT", "20")
	t.Setenv("LOG_LEVEL", "error")
	cfg, err := config.Load(t.TempDir() + "/.env")
	if err != nil {
		t.Fatal(err)
	}
	application, err := app.New(cfg, "v", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	st, ct := sdk.NewInMemoryTransports()
	ss, err := application.Server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	cs, err := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "v"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()
	listed, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertApprovedTools(listed.Tools); err != nil {
		t.Fatal(err)
	}
	prompts, err := cs.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(prompts.Prompts) != 0 {
		t.Error("server unexpectedly exposes prompts")
	}
	resources, err := cs.ListResources(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(resources.Resources) != 0 {
		t.Error("server unexpectedly exposes resources")
	}
}

func TestRegistryPolicyRejectsDeliberatelyAddedExtraTool(t *testing.T) {
	destructive := false
	err := assertApprovedTools([]*sdk.Tool{{Name: approvedToolNames[0], Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: &destructive}}, {Name: "kaiten_extra", Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: &destructive}}})
	if err == nil {
		t.Fatal("policy accepted a deliberately added extra registration")
	}
}

func assertApprovedTools(tools []*sdk.Tool) error {
	got := make([]string, 0, len(tools))
	for _, tool := range tools {
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint || tool.Annotations.DestructiveHint == nil || *tool.Annotations.DestructiveHint {
			return fmt.Errorf("tool %q is not read-only and non-destructive", tool.Name)
		}
		got = append(got, tool.Name)
	}
	want := append([]string(nil), approvedToolNames...)
	sort.Strings(want)
	sort.Strings(got)
	if len(got) != len(want) {
		return fmt.Errorf("tool count = %d, want %d (%v)", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			return fmt.Errorf("tool catalog = %v, want %v", got, want)
		}
	}
	return nil
}
