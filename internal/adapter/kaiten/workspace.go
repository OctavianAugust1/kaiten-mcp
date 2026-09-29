package kaiten

import (
	"context"
	"net/url"
	"strconv"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

// ListSpaces reads one documented offset page of spaces.
func (c *Client) ListSpaces(ctx context.Context, page domain.PageRequest) (domain.Result[[]domain.Object], error) {
	query := url.Values{
		"limit":  {strconv.Itoa(page.Limit)},
		"offset": {strconv.Itoa(page.Offset)},
	}
	return c.list(ctx, "spaces", query, &page)
}

// GetSpace reads one space by its validated numeric identifier.
func (c *Client) GetSpace(ctx context.Context, spaceID domain.ID) (domain.Result[domain.Object], error) {
	return c.detail(ctx, "spaces/"+strconv.FormatInt(int64(spaceID), 10))
}

// ListBoards reads the boards in one space.
func (c *Client) ListBoards(ctx context.Context, spaceID domain.ID) (domain.Result[[]domain.Object], error) {
	return c.list(ctx, "spaces/"+strconv.FormatInt(int64(spaceID), 10)+"/boards", nil, nil)
}

// GetBoard reads one board by its validated numeric identifier.
func (c *Client) GetBoard(ctx context.Context, boardID domain.ID) (domain.Result[domain.Object], error) {
	return c.detail(ctx, "boards/"+strconv.FormatInt(int64(boardID), 10))
}

// ListColumns reads the columns in one board.
func (c *Client) ListColumns(ctx context.Context, boardID domain.ID) (domain.Result[[]domain.Object], error) {
	return c.list(ctx, "boards/"+strconv.FormatInt(int64(boardID), 10)+"/columns", nil, nil)
}

// ListSubcolumns reads the subcolumns in one column.
func (c *Client) ListSubcolumns(ctx context.Context, columnID domain.ID) (domain.Result[[]domain.Object], error) {
	return c.list(ctx, "columns/"+strconv.FormatInt(int64(columnID), 10)+"/subcolumns", nil, nil)
}

// ListLanes reads the lanes in one board.
func (c *Client) ListLanes(ctx context.Context, boardID domain.ID) (domain.Result[[]domain.Object], error) {
	return c.list(ctx, "boards/"+strconv.FormatInt(int64(boardID), 10)+"/lanes", nil, nil)
}

func (c *Client) list(
	ctx context.Context,
	endpoint string,
	query url.Values,
	page *domain.PageRequest,
) (domain.Result[[]domain.Object], error) {
	var data []domain.Object
	if _, err := c.getJSON(ctx, endpoint, query, &data); err != nil {
		return domain.Result[[]domain.Object]{}, err
	}
	result := domain.Result[[]domain.Object]{Data: data}
	if page != nil {
		offset := page.Offset
		pagination := &domain.Pagination{Limit: page.Limit, Offset: &offset}
		if len(data) == page.Limit {
			nextOffset := page.Offset + page.Limit
			pagination.NextOffset = &nextOffset
		}
		result.Pagination = pagination
	}
	return result, nil
}

func (c *Client) detail(ctx context.Context, endpoint string) (domain.Result[domain.Object], error) {
	var data domain.Object
	if _, err := c.getJSON(ctx, endpoint, nil, &data); err != nil {
		return domain.Result[domain.Object]{}, err
	}
	return domain.Result[domain.Object]{Data: data}, nil
}
