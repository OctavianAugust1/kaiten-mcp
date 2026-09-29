package user

import (
	"context"
	"testing"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

func TestServicePropagatesUserReads(t *testing.T) {
	t.Parallel()
	stub := &readerStub{}
	service := NewService(stub)
	page, _ := domain.NewPage(20, 0, 50)
	if _, err := service.ListUsers(context.Background(), domain.UserQuery{Page: page}); err != nil {
		t.Fatal(err)
	}
	spaceID, _ := domain.NewID(7)
	userID, _ := domain.NewID(8)
	if _, err := service.GetCurrentUser(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ListSpaceUsers(context.Background(), spaceID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetSpaceUser(context.Background(), spaceID, userID); err != nil {
		t.Fatal(err)
	}
	if stub.calls != 4 {
		t.Errorf("calls = %d, want 4", stub.calls)
	}
}

type readerStub struct{ calls int }

func (s *readerStub) ListUsers(context.Context, domain.UserQuery) (domain.Result[[]domain.Object], error) {
	s.calls++
	return domain.Result[[]domain.Object]{}, nil
}
func (s *readerStub) GetCurrentUser(context.Context) (domain.Result[domain.Object], error) {
	s.calls++
	return domain.Result[domain.Object]{}, nil
}
func (s *readerStub) ListSpaceUsers(context.Context, domain.ID) (domain.Result[[]domain.Object], error) {
	s.calls++
	return domain.Result[[]domain.Object]{}, nil
}
func (s *readerStub) GetSpaceUser(context.Context, domain.ID, domain.ID) (domain.Result[domain.Object], error) {
	s.calls++
	return domain.Result[domain.Object]{}, nil
}
