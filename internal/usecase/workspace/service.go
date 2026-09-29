package workspace

import (
	"context"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

// Reader is the workspace read port consumed by this use case.
type Reader interface {
	ListSpaces(context.Context, domain.PageRequest) (domain.Result[[]domain.Object], error)
	GetSpace(context.Context, domain.ID) (domain.Result[domain.Object], error)
	ListBoards(context.Context, domain.ID) (domain.Result[[]domain.Object], error)
	GetBoard(context.Context, domain.ID) (domain.Result[domain.Object], error)
	ListColumns(context.Context, domain.ID) (domain.Result[[]domain.Object], error)
	ListSubcolumns(context.Context, domain.ID) (domain.Result[[]domain.Object], error)
	ListLanes(context.Context, domain.ID) (domain.Result[[]domain.Object], error)
}

// Service coordinates workspace reads without depending on HTTP or MCP.
type Service struct {
	reader Reader
}

// NewService injects the workspace reader explicitly.
func NewService(reader Reader) *Service {
	return &Service{reader: reader}
}

func (service *Service) ListSpaces(ctx context.Context, page domain.PageRequest) (domain.Result[[]domain.Object], error) {
	return service.reader.ListSpaces(ctx, page)
}

func (service *Service) GetSpace(ctx context.Context, id domain.ID) (domain.Result[domain.Object], error) {
	return service.reader.GetSpace(ctx, id)
}

func (service *Service) ListBoards(ctx context.Context, id domain.ID) (domain.Result[[]domain.Object], error) {
	return service.reader.ListBoards(ctx, id)
}

func (service *Service) GetBoard(ctx context.Context, id domain.ID) (domain.Result[domain.Object], error) {
	return service.reader.GetBoard(ctx, id)
}

func (service *Service) ListColumns(ctx context.Context, id domain.ID) (domain.Result[[]domain.Object], error) {
	return service.reader.ListColumns(ctx, id)
}

func (service *Service) ListSubcolumns(ctx context.Context, id domain.ID) (domain.Result[[]domain.Object], error) {
	return service.reader.ListSubcolumns(ctx, id)
}

func (service *Service) ListLanes(ctx context.Context, id domain.ID) (domain.Result[[]domain.Object], error) {
	return service.reader.ListLanes(ctx, id)
}
