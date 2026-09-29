package document

import (
	"context"
	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
	"testing"
)

func TestServicePropagatesDocumentReads(t *testing.T) {
	t.Parallel()
	stub := &readerStub{}
	service := NewService(stub, 50)
	if _, err := service.ListDocuments(context.Background(), domain.DocumentQuery{Limit: 20}); err != nil {
		t.Fatal(err)
	}
	uid, _ := domain.ParseUID("550e8400-e29b-41d4-a716-446655440000")
	id, _ := domain.NewID(7)
	if _, err := service.GetDocument(context.Background(), uid); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetDocumentSchema(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ListDocumentGroups(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetDocumentGroup(context.Background(), uid); err != nil {
		t.Fatal(err)
	}
	if stub.defaultLimit != 50 || stub.calls != 5 {
		t.Errorf("default/calls = %d/%d, want 50/5", stub.defaultLimit, stub.calls)
	}
}

type readerStub struct{ calls, defaultLimit int }

func (s *readerStub) ListDocuments(context.Context, domain.DocumentQuery, int) (domain.Result[[]domain.Object], error) {
	s.calls++
	s.defaultLimit = 50
	return domain.Result[[]domain.Object]{}, nil
}
func (s *readerStub) GetDocument(context.Context, domain.UID) (domain.Result[domain.Object], error) {
	s.calls++
	return domain.Result[domain.Object]{}, nil
}
func (s *readerStub) GetDocumentSchema(context.Context, domain.ID) (domain.Result[domain.Object], error) {
	s.calls++
	return domain.Result[domain.Object]{}, nil
}
func (s *readerStub) ListDocumentGroups(context.Context) (domain.Result[[]domain.Object], error) {
	s.calls++
	return domain.Result[[]domain.Object]{}, nil
}
func (s *readerStub) GetDocumentGroup(context.Context, domain.UID) (domain.Result[domain.Object], error) {
	s.calls++
	return domain.Result[domain.Object]{}, nil
}
