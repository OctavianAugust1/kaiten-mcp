package user

import (
	"context"
	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

type Reader interface {
	ListUsers(context.Context, domain.UserQuery) (domain.Result[[]domain.Object], error)
	GetCurrentUser(context.Context) (domain.Result[domain.Object], error)
	ListSpaceUsers(context.Context, domain.ID) (domain.Result[[]domain.Object], error)
	GetSpaceUser(context.Context, domain.ID, domain.ID) (domain.Result[domain.Object], error)
}
type Service struct{ reader Reader }

func NewService(reader Reader) *Service { return &Service{reader: reader} }
func (s *Service) ListUsers(ctx context.Context, q domain.UserQuery) (domain.Result[[]domain.Object], error) {
	return s.reader.ListUsers(ctx, q)
}
func (s *Service) GetCurrentUser(ctx context.Context) (domain.Result[domain.Object], error) {
	return s.reader.GetCurrentUser(ctx)
}
func (s *Service) ListSpaceUsers(ctx context.Context, id domain.ID) (domain.Result[[]domain.Object], error) {
	return s.reader.ListSpaceUsers(ctx, id)
}
func (s *Service) GetSpaceUser(ctx context.Context, spaceID, userID domain.ID) (domain.Result[domain.Object], error) {
	return s.reader.GetSpaceUser(ctx, spaceID, userID)
}
