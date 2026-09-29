package card

import (
	"context"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

// Reader is the card read port consumed by this use case.
type Reader interface {
	ListCards(context.Context, domain.CardQuery, int) (domain.Result[[]domain.Object], error)
	GetCard(context.Context, domain.ID) (domain.Result[domain.Object], error)
	ListCardComments(context.Context, domain.ID) (domain.Result[[]domain.Object], error)
	ListCardMembers(context.Context, domain.ID) (domain.Result[[]domain.Object], error)
}

// Service coordinates card reads without depending on the HTTP adapter.
type Service struct {
	reader           Reader
	defaultPageLimit int
}

// NewService injects the card reader and configured default page limit.
func NewService(reader Reader, defaultPageLimit int) *Service {
	return &Service{reader: reader, defaultPageLimit: defaultPageLimit}
}

func (service *Service) ListCards(ctx context.Context, query domain.CardQuery) (domain.Result[[]domain.Object], error) {
	return service.reader.ListCards(ctx, query, service.defaultPageLimit)
}

func (service *Service) GetCard(ctx context.Context, id domain.ID) (domain.Result[domain.Object], error) {
	return service.reader.GetCard(ctx, id)
}

func (service *Service) ListCardComments(ctx context.Context, id domain.ID) (domain.Result[[]domain.Object], error) {
	return service.reader.ListCardComments(ctx, id)
}

func (service *Service) ListCardMembers(ctx context.Context, id domain.ID) (domain.Result[[]domain.Object], error) {
	return service.reader.ListCardMembers(ctx, id)
}
