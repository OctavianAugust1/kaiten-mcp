package card

import (
	"context"
	"errors"
	"testing"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

func TestServicePropagatesCardReadsAndTypedUpstreamErrors(t *testing.T) {
	t.Parallel()

	upstream := &domain.Error{Kind: domain.ErrorNotFound, Operation: "Kaiten GET", StatusCode: 404}
	reader := &fakeReader{err: upstream}
	service := NewService(reader, 50)
	id, _ := domain.NewID(7)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.GetCard(ctx, id)
	if !errors.Is(err, upstream) {
		t.Fatalf("GetCard() error = %v, want typed upstream error", err)
	}
	if reader.context != ctx {
		t.Fatal("GetCard() did not pass the caller context to the reader")
	}
	if _, err := service.ListCards(context.Background(), domain.CardQuery{Limit: 20}); err != nil {
		t.Fatalf("ListCards() error = %v", err)
	}
	if reader.defaultLimit != 50 || reader.query.Limit != 20 {
		t.Errorf("ListCards() query/default = %#v/%d, want limit 20/default 50", reader.query, reader.defaultLimit)
	}
	if result, err := service.ListCardComments(context.Background(), id); err != nil || result.Data[0]["kind"] != "comments" {
		t.Errorf("ListCardComments() result/error = %#v/%v", result, err)
	}
	if result, err := service.ListCardMembers(context.Background(), id); err != nil || result.Data[0]["kind"] != "members" {
		t.Errorf("ListCardMembers() result/error = %#v/%v", result, err)
	}
}

type fakeReader struct {
	context      context.Context
	query        domain.CardQuery
	defaultLimit int
	err          error
}

func (reader *fakeReader) ListCards(ctx context.Context, query domain.CardQuery, defaultLimit int) (domain.Result[[]domain.Object], error) {
	reader.context, reader.query, reader.defaultLimit = ctx, query, defaultLimit
	return domain.Result[[]domain.Object]{Data: []domain.Object{{"kind": "cards"}}}, nil
}

func (reader *fakeReader) GetCard(ctx context.Context, _ domain.ID) (domain.Result[domain.Object], error) {
	reader.context = ctx
	return domain.Result[domain.Object]{}, reader.err
}

func (reader *fakeReader) ListCardComments(context.Context, domain.ID) (domain.Result[[]domain.Object], error) {
	return domain.Result[[]domain.Object]{Data: []domain.Object{{"kind": "comments"}}}, nil
}

func (reader *fakeReader) ListCardMembers(context.Context, domain.ID) (domain.Result[[]domain.Object], error) {
	return domain.Result[[]domain.Object]{Data: []domain.Object{{"kind": "members"}}}, nil
}
