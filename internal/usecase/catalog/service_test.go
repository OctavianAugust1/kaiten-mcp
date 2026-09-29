package catalog

import (
	"context"
	"testing"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

func TestServicePropagatesCatalogReads(t *testing.T) {
	t.Parallel()
	stub := &readerStub{}
	service := NewService(stub)
	id, _ := domain.NewID(9)
	if _, err := service.ListCardTypes(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetCardType(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ListTags(context.Background()); err != nil {
		t.Fatal(err)
	}
	if stub.calls != 3 {
		t.Errorf("calls = %d, want 3", stub.calls)
	}
}

type readerStub struct{ calls int }

func (s *readerStub) ListCardTypes(context.Context) (domain.Result[[]domain.Object], error) {
	s.calls++
	return domain.Result[[]domain.Object]{}, nil
}
func (s *readerStub) GetCardType(context.Context, domain.ID) (domain.Result[domain.Object], error) {
	s.calls++
	return domain.Result[domain.Object]{}, nil
}
func (s *readerStub) ListTags(context.Context) (domain.Result[[]domain.Object], error) {
	s.calls++
	return domain.Result[[]domain.Object]{}, nil
}
