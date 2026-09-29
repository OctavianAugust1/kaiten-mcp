package kaiten

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

func cardQueryValues(query domain.CardQuery, defaultLimit int) (url.Values, domain.SearchPage, error) {
	page, err := domain.NewSearchPage(query.Version, query.Limit, query.Offset, query.StartPosition, defaultLimit)
	if err != nil {
		return nil, domain.SearchPage{}, err
	}
	if query.IncludeSearchPreview != nil && page.Version != domain.SearchVersion2 {
		return nil, domain.SearchPage{}, errors.New("include_search_preview requires search version 2")
	}
	values := make(url.Values)
	dates := []struct{ key, value string }{
		{"created_before", query.CreatedBefore}, {"created_after", query.CreatedAfter},
		{"updated_before", query.UpdatedBefore}, {"updated_after", query.UpdatedAfter},
		{"first_moved_in_progress_after", query.FirstMovedInProgressAfter}, {"first_moved_in_progress_before", query.FirstMovedInProgressBefore},
		{"last_moved_to_done_at_after", query.LastMovedToDoneAfter}, {"last_moved_to_done_at_before", query.LastMovedToDoneBefore},
		{"due_date_after", query.DueDateAfter}, {"due_date_before", query.DueDateBefore},
	}
	for _, field := range dates {
		if field.value == "" {
			continue
		}
		if _, err := time.Parse(time.RFC3339Nano, field.value); err != nil {
			return nil, domain.SearchPage{}, errors.New(field.key + " must use ISO 8601/RFC3339 format")
		}
		values.Set(field.key, field.value)
	}
	setString(values, "query", query.Query)
	setString(values, "tag", query.Tag)
	setString(values, "external_id", query.ExternalID)
	if query.Version != 0 {
		values.Set("version", strconv.Itoa(query.Version))
	}
	values.Set("limit", strconv.Itoa(page.Limit))
	if page.Version == domain.SearchVersion1 {
		values.Set("offset", strconv.Itoa(page.Offset))
	} else {
		setString(values, "start_position", page.StartPosition)
	}

	for _, field := range []struct {
		key   string
		value []int64
	}{
		{"tag_ids", query.TagIDs}, {"type_ids", query.TypeIDs}, {"exclude_board_ids", query.ExcludeBoardIDs},
		{"exclude_lane_ids", query.ExcludeLaneIDs}, {"exclude_column_ids", query.ExcludeColumnIDs}, {"column_ids", query.ColumnIDs},
		{"member_ids", query.MemberIDs}, {"owner_ids", query.OwnerIDs}, {"responsible_ids", query.ResponsibleIDs},
		{"exclude_owner_ids", query.ExcludeOwnerIDs}, {"exclude_card_ids", query.ExcludeCardIDs}, {"organizations_ids", query.OrganizationIDs},
	} {
		encoded, err := positiveIDs(field.value)
		if err != nil {
			return nil, domain.SearchPage{}, errors.New(field.key + " must contain positive identifiers")
		}
		setString(values, field.key, encoded)
	}
	if len(query.States) > 0 {
		parts := make([]string, len(query.States))
		for index, state := range query.States {
			if state < 1 || state > 3 {
				return nil, domain.SearchPage{}, errors.New("states must contain values from 1 to 3")
			}
			parts[index] = strconv.Itoa(state)
		}
		values.Set("states", strings.Join(parts, ","))
	}
	for _, field := range []struct {
		key   string
		value int64
	}{
		{"space_id", query.SpaceID}, {"order_space_id", query.OrderSpaceID}, {"board_id", query.BoardID},
		{"column_id", query.ColumnID}, {"lane_id", query.LaneID}, {"type_id", query.TypeID},
		{"responsible_id", query.ResponsibleID}, {"owner_id", query.OwnerID},
	} {
		if field.value < 0 {
			return nil, domain.SearchPage{}, errors.New(field.key + " must be positive")
		}
		if field.value > 0 {
			values.Set(field.key, strconv.FormatInt(field.value, 10))
		}
	}
	if query.Condition < 0 || query.Condition > 2 {
		return nil, domain.SearchPage{}, errors.New("condition must be 1 or 2")
	}
	if query.Condition > 0 {
		values.Set("condition", strconv.Itoa(query.Condition))
	}
	if err := setNonEmptyCSV(values, "additional_card_fields", query.AdditionalCardFields); err != nil {
		return nil, domain.SearchPage{}, err
	}
	if err := setNonEmptyCSV(values, "search_fields", query.SearchFields); err != nil {
		return nil, domain.SearchPage{}, err
	}
	if err := setNonEmptyCSV(values, "order_by", query.OrderBy); err != nil {
		return nil, domain.SearchPage{}, err
	}
	for _, direction := range query.OrderDirection {
		if direction != "asc" && direction != "desc" {
			return nil, domain.SearchPage{}, errors.New("order_direction values must be asc or desc")
		}
	}
	if len(query.OrderDirection) > 0 && len(query.OrderBy) != len(query.OrderDirection) {
		return nil, domain.SearchPage{}, errors.New("order_direction must correspond to order_by")
	}
	if err := setNonEmptyCSV(values, "order_direction", query.OrderDirection); err != nil {
		return nil, domain.SearchPage{}, err
	}
	setBool(values, "include_search_preview", query.IncludeSearchPreview)
	setBool(values, "archived", query.Archived)
	setBool(values, "asap", query.ASAP)
	setBool(values, "overdue", query.Overdue)
	setBool(values, "done_on_time", query.DoneOnTime)
	setBool(values, "with_due_date", query.WithDueDate)
	setBool(values, "is_request", query.IsRequest)
	setBool(values, "broken_api", query.BrokenAPI)
	return values, page, nil
}

func positiveIDs(ids []int64) (string, error) {
	parts := make([]string, len(ids))
	for index, id := range ids {
		if id <= 0 {
			return "", errors.New("identifier must be positive")
		}
		parts[index] = strconv.FormatInt(id, 10)
	}
	return strings.Join(parts, ","), nil
}

func setString(values url.Values, key, value string) {
	if value != "" {
		values.Set(key, value)
	}
}

func setBool(values url.Values, key string, value *bool) {
	if value != nil {
		values.Set(key, strconv.FormatBool(*value))
	}
}

func setNonEmptyCSV(values url.Values, key string, items []string) error {
	for _, item := range items {
		if item == "" || strings.Contains(item, ",") {
			return errors.New(key + " must contain non-empty comma-free values")
		}
	}
	if len(items) > 0 {
		values.Set(key, strings.Join(items, ","))
	}
	return nil
}
