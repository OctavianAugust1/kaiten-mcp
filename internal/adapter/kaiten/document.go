package kaiten

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

func (c *Client) ListDocuments(ctx context.Context, query domain.DocumentQuery, defaultLimit int) (domain.Result[[]domain.Object], error) {
	page, err := domain.NewSearchPage(query.Version, query.Limit, query.Offset, query.StartPosition, defaultLimit)
	if err != nil {
		return domain.Result[[]domain.Object]{}, &domain.Error{Kind: domain.ErrorInvalidArgument, Operation: "validate document query", Cause: err}
	}
	if query.IncludeSearchPreview != nil && page.Version != domain.SearchVersion2 {
		return domain.Result[[]domain.Object]{}, &domain.Error{Kind: domain.ErrorInvalidArgument, Operation: "validate document query", Cause: errors.New("include_search_preview requires version 2")}
	}
	values := url.Values{"limit": {strconv.Itoa(page.Limit)}}
	setString(values, "query", query.Query)
	setString(values, "condition", query.Condition)
	if err := setNonEmptyCSV(values, "fields", query.Fields); err != nil {
		return domain.Result[[]domain.Object]{}, &domain.Error{Kind: domain.ErrorInvalidArgument, Operation: "validate document query", Cause: err}
	}
	if query.Version != 0 {
		values.Set("version", strconv.Itoa(query.Version))
	}
	if page.Version == domain.SearchVersion1 {
		values.Set("offset", strconv.Itoa(page.Offset))
		return c.list(ctx, "documents", values, &page.PageRequest)
	}
	setString(values, "start_position", page.StartPosition)
	setBool(values, "include_search_preview", query.IncludeSearchPreview)
	var response struct {
		Result   []domain.Object `json:"result"`
		Position string          `json:"position"`
	}
	if _, err := c.getJSON(ctx, "documents", values, &response); err != nil {
		return domain.Result[[]domain.Object]{}, err
	}
	return domain.Result[[]domain.Object]{Data: response.Result, Pagination: &domain.Pagination{Limit: page.Limit, NextPosition: response.Position}}, nil
}

func (c *Client) GetDocument(ctx context.Context, uid domain.UID) (domain.Result[domain.Object], error) {
	return c.detail(ctx, "documents/"+uid.String())
}
func (c *Client) GetDocumentSchema(ctx context.Context, id domain.ID) (domain.Result[domain.Object], error) {
	return c.detail(ctx, "document-schemas/"+strconv.FormatInt(int64(id), 10))
}
func (c *Client) ListDocumentGroups(ctx context.Context) (domain.Result[[]domain.Object], error) {
	return c.list(ctx, "document-groups", nil, nil)
}
func (c *Client) GetDocumentGroup(ctx context.Context, uid domain.UID) (domain.Result[domain.Object], error) {
	return c.detail(ctx, "document-groups/"+uid.String())
}
