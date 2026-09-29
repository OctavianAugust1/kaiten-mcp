package kaiten

import (
	"reflect"
	"testing"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

func TestCardQueryValuesEncodesEveryDocumentedNonBetaFilter(t *testing.T) {
	t.Parallel()

	trueValue := true
	falseValue := false
	query := domain.CardQuery{
		CreatedBefore: "2026-01-02T03:04:05Z", CreatedAfter: "2025-01-02T03:04:05Z",
		UpdatedBefore: "2026-02-02T03:04:05Z", UpdatedAfter: "2025-02-02T03:04:05Z",
		FirstMovedInProgressAfter: "2025-03-02T03:04:05Z", FirstMovedInProgressBefore: "2026-03-02T03:04:05Z",
		LastMovedToDoneAfter: "2025-04-02T03:04:05Z", LastMovedToDoneBefore: "2026-04-02T03:04:05Z",
		DueDateAfter: "2025-05-02T03:04:05Z", DueDateBefore: "2026-05-02T03:04:05Z",
		Query: "urgent & safe", Version: 2, Tag: "backend", TagIDs: []int64{1, 2}, TypeIDs: []int64{3},
		ExcludeBoardIDs: []int64{4}, ExcludeLaneIDs: []int64{5}, ExcludeColumnIDs: []int64{6},
		ColumnIDs: []int64{7}, MemberIDs: []int64{8}, OwnerIDs: []int64{9}, ResponsibleIDs: []int64{10},
		States: []int{1, 3}, ExternalID: "EXT-1", AdditionalCardFields: []string{"description"}, SearchFields: []string{"title"},
		SpaceID: 11, Limit: 20, StartPosition: "cursor", IncludeSearchPreview: &trueValue, OrderSpaceID: 12,
		BoardID: 13, ColumnID: 14, LaneID: 15, Condition: 1, TypeID: 16, ResponsibleID: 17, OwnerID: 18,
		Archived: &falseValue, ASAP: &trueValue, Overdue: &falseValue, DoneOnTime: &trueValue, WithDueDate: &falseValue,
		OrderBy: []string{"updated", "id"}, OrderDirection: []string{"desc", "asc"}, IsRequest: &falseValue,
		ExcludeOwnerIDs: []int64{19}, ExcludeCardIDs: []int64{20}, OrganizationIDs: []int64{21}, BrokenAPI: &falseValue,
	}

	values, page, err := cardQueryValues(query, 50)
	if err != nil {
		t.Fatalf("Values() error = %v", err)
	}
	if page.Version != 2 || page.StartPosition != "cursor" || page.Limit != 20 {
		t.Errorf("page = %#v, want version 2 cursor limit 20", page)
	}
	want := map[string]string{
		"created_before": "2026-01-02T03:04:05Z", "created_after": "2025-01-02T03:04:05Z",
		"updated_before": "2026-02-02T03:04:05Z", "updated_after": "2025-02-02T03:04:05Z",
		"first_moved_in_progress_after": "2025-03-02T03:04:05Z", "first_moved_in_progress_before": "2026-03-02T03:04:05Z",
		"last_moved_to_done_at_after": "2025-04-02T03:04:05Z", "last_moved_to_done_at_before": "2026-04-02T03:04:05Z",
		"due_date_after": "2025-05-02T03:04:05Z", "due_date_before": "2026-05-02T03:04:05Z",
		"query": "urgent & safe", "version": "2", "tag": "backend", "tag_ids": "1,2", "type_ids": "3",
		"exclude_board_ids": "4", "exclude_lane_ids": "5", "exclude_column_ids": "6", "column_ids": "7",
		"member_ids": "8", "owner_ids": "9", "responsible_ids": "10", "states": "1,3", "external_id": "EXT-1",
		"additional_card_fields": "description", "search_fields": "title", "space_id": "11", "limit": "20",
		"start_position": "cursor", "include_search_preview": "true", "order_space_id": "12", "board_id": "13",
		"column_id": "14", "lane_id": "15", "condition": "1", "type_id": "16", "responsible_id": "17",
		"owner_id": "18", "archived": "false", "asap": "true", "overdue": "false", "done_on_time": "true",
		"with_due_date": "false", "order_by": "updated,id", "order_direction": "desc,asc", "is_request": "false",
		"exclude_owner_ids": "19", "exclude_card_ids": "20", "organizations_ids": "21", "broken_api": "false",
	}
	if got := values.Get("filter"); got != "" {
		t.Fatalf("beta filter unexpectedly encoded: %q", got)
	}
	for key, wantValue := range want {
		if got := values.Get(key); got != wantValue {
			t.Errorf("query[%s] = %q, want %q", key, got, wantValue)
		}
	}
	if len(values) != len(want) {
		t.Errorf("encoded keys = %d, want %d: %v", len(values), len(want), values)
	}
}

func TestCardQueryValuesRejectsInvalidCombinations(t *testing.T) {
	t.Parallel()

	tests := []domain.CardQuery{
		{Version: 1, StartPosition: "cursor"},
		{Version: 2, Offset: 1},
		{Version: 3},
		{Limit: 101},
		{Offset: -1},
		{TagIDs: []int64{0}},
		{States: []int{4}},
		{Condition: 3},
		{CreatedAfter: "not-a-date"},
		{OrderDirection: []string{"sideways"}},
	}
	for _, query := range tests {
		if _, _, err := cardQueryValues(query, 50); err == nil {
			t.Errorf("Values(%#v) unexpectedly succeeded", query)
		}
	}
}

func TestCardQueryHasNoBetaFilterField(t *testing.T) {
	t.Parallel()
	if _, exists := reflect.TypeFor[domain.CardQuery]().FieldByName("Filter"); exists {
		t.Fatal("CardQuery exposes the beta filter field")
	}
}

func FuzzCardQueryTextIsEncodedAsData(f *testing.F) {
	for _, seed := range []string{"urgent", "x&filter=DELETE", "../cards", "кириллица"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		values, _, err := cardQueryValues(domain.CardQuery{Query: text}, 50)
		if err != nil {
			t.Fatalf("Values() error = %v", err)
		}
		if values.Get("query") != text || values.Get("filter") != "" {
			t.Fatalf("values = %v, want query as data and no beta filter", values)
		}
	})
}
