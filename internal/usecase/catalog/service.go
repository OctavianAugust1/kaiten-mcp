package catalog

import (
	"context"
	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

type Reader interface {
	ListCardTypes(context.Context) (domain.Result[[]domain.Object], error)
	GetCardType(context.Context, domain.ID) (domain.Result[domain.Object], error)
	ListTags(context.Context) (domain.Result[[]domain.Object], error)
}
type Service struct{ reader Reader }

func NewService(reader Reader) *Service { return &Service{reader: reader} }
func (s *Service) ListCardTypes(ctx context.Context) (domain.Result[[]domain.Object], error) {
	return s.reader.ListCardTypes(ctx)
}
func (s *Service) GetCardType(ctx context.Context, id domain.ID) (domain.Result[domain.Object], error) {
	return s.reader.GetCardType(ctx, id)
}
func (s *Service) ListTags(ctx context.Context) (domain.Result[[]domain.Object], error) {
	return s.reader.ListTags(ctx)
}
