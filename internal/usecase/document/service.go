package document

import (
	"context"
	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

type Reader interface {
	ListDocuments(context.Context, domain.DocumentQuery, int) (domain.Result[[]domain.Object], error)
	GetDocument(context.Context, domain.UID) (domain.Result[domain.Object], error)
	GetDocumentSchema(context.Context, domain.ID) (domain.Result[domain.Object], error)
	ListDocumentGroups(context.Context) (domain.Result[[]domain.Object], error)
	GetDocumentGroup(context.Context, domain.UID) (domain.Result[domain.Object], error)
}
type Service struct {
	reader           Reader
	defaultPageLimit int
}

func NewService(reader Reader, defaultPageLimit int) *Service {
	return &Service{reader: reader, defaultPageLimit: defaultPageLimit}
}
func (s *Service) ListDocuments(ctx context.Context, q domain.DocumentQuery) (domain.Result[[]domain.Object], error) {
	return s.reader.ListDocuments(ctx, q, s.defaultPageLimit)
}
func (s *Service) GetDocument(ctx context.Context, uid domain.UID) (domain.Result[domain.Object], error) {
	return s.reader.GetDocument(ctx, uid)
}
func (s *Service) GetDocumentSchema(ctx context.Context, id domain.ID) (domain.Result[domain.Object], error) {
	return s.reader.GetDocumentSchema(ctx, id)
}
func (s *Service) ListDocumentGroups(ctx context.Context) (domain.Result[[]domain.Object], error) {
	return s.reader.ListDocumentGroups(ctx)
}
func (s *Service) GetDocumentGroup(ctx context.Context, uid domain.UID) (domain.Result[domain.Object], error) {
	return s.reader.GetDocumentGroup(ctx, uid)
}
