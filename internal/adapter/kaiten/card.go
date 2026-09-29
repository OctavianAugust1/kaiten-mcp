package kaiten

import (
	"context"
	"strconv"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

// ListCards reads exactly one validated card search page.
func (c *Client) ListCards(ctx context.Context, query domain.CardQuery, defaultLimit int) (domain.Result[[]domain.Object], error) {
	values, page, err := cardQueryValues(query, defaultLimit)
	if err != nil {
		return domain.Result[[]domain.Object]{}, &domain.Error{Kind: domain.ErrorInvalidArgument, Operation: "validate card query", Cause: err}
	}
	if page.Version == domain.SearchVersion1 {
		return c.list(ctx, "cards", values, &page.PageRequest)
	}
	var response struct {
		Result   []domain.Object `json:"result"`
		Position string          `json:"position"`
	}
	if _, err := c.getJSON(ctx, "cards", values, &response); err != nil {
		return domain.Result[[]domain.Object]{}, err
	}
	return domain.Result[[]domain.Object]{
		Data: response.Result,
		Pagination: &domain.Pagination{
			Limit:        page.Limit,
			NextPosition: response.Position,
		},
	}, nil
}

// GetCard reads one card by ID.
func (c *Client) GetCard(ctx context.Context, cardID domain.ID) (domain.Result[domain.Object], error) {
	return c.detail(ctx, "cards/"+strconv.FormatInt(int64(cardID), 10))
}

// ListCardComments reads the comments attached to one card.
func (c *Client) ListCardComments(ctx context.Context, cardID domain.ID) (domain.Result[[]domain.Object], error) {
	return c.list(ctx, "cards/"+strconv.FormatInt(int64(cardID), 10)+"/comments", nil, nil)
}

// ListCardMembers reads the members attached to one card.
func (c *Client) ListCardMembers(ctx context.Context, cardID domain.ID) (domain.Result[[]domain.Object], error) {
	return c.list(ctx, "cards/"+strconv.FormatInt(int64(cardID), 10)+"/members", nil, nil)
}
