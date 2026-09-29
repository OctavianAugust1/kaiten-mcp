package workspace

import (
	"context"
	"testing"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

func TestServicePropagatesWorkspaceReadsAndPagination(t *testing.T) {
	t.Parallel()

	type contextKeyType struct{}
	contextKey := contextKeyType{}
	ctx := context.WithValue(context.Background(), contextKey, "marker")
	page := domain.PageRequest{Limit: 20, Offset: 5}
	reader := &fakeReader{t: t, expectedContext: ctx, expectedPage: page}
	service := NewService(reader)
	id, _ := domain.NewID(7)

	spaces, err := service.ListSpaces(ctx, page)
	if err != nil || spaces.Pagination == nil || spaces.Pagination.Limit != 20 {
		t.Fatalf("ListSpaces() result/error = %#v/%v", spaces, err)
	}
	for name, call := range map[string]func() (domain.Result[[]domain.Object], error){
		"boards":     func() (domain.Result[[]domain.Object], error) { return service.ListBoards(ctx, id) },
		"columns":    func() (domain.Result[[]domain.Object], error) { return service.ListColumns(ctx, id) },
		"subcolumns": func() (domain.Result[[]domain.Object], error) { return service.ListSubcolumns(ctx, id) },
		"lanes":      func() (domain.Result[[]domain.Object], error) { return service.ListLanes(ctx, id) },
	} {
		result, err := call()
		if err != nil || result.Data[0]["kind"] != name {
			t.Errorf("%s result/error = %#v/%v", name, result, err)
		}
	}
	if result, err := service.GetSpace(ctx, id); err != nil || result.Data["kind"] != "space" {
		t.Errorf("GetSpace() result/error = %#v/%v", result, err)
	}
	if result, err := service.GetBoard(ctx, id); err != nil || result.Data["kind"] != "board" {
		t.Errorf("GetBoard() result/error = %#v/%v", result, err)
	}
}

type fakeReader struct {
	t               *testing.T
	expectedContext context.Context
	expectedPage    domain.PageRequest
}

func (reader *fakeReader) checkContext(ctx context.Context) {
	reader.t.Helper()
	if ctx != reader.expectedContext {
		reader.t.Fatal("use case replaced the caller context")
	}
}

func (reader *fakeReader) ListSpaces(ctx context.Context, page domain.PageRequest) (domain.Result[[]domain.Object], error) {
	reader.checkContext(ctx)
	if page != reader.expectedPage {
		reader.t.Errorf("page = %#v, want %#v", page, reader.expectedPage)
	}
	return domain.Result[[]domain.Object]{Data: []domain.Object{{"kind": "spaces"}}, Pagination: &domain.Pagination{Limit: page.Limit}}, nil
}

func (reader *fakeReader) GetSpace(ctx context.Context, _ domain.ID) (domain.Result[domain.Object], error) {
	reader.checkContext(ctx)
	return domain.Result[domain.Object]{Data: domain.Object{"kind": "space"}}, nil
}

func (reader *fakeReader) ListBoards(ctx context.Context, _ domain.ID) (domain.Result[[]domain.Object], error) {
	reader.checkContext(ctx)
	return listResult("boards"), nil
}

func (reader *fakeReader) GetBoard(ctx context.Context, _ domain.ID) (domain.Result[domain.Object], error) {
	reader.checkContext(ctx)
	return domain.Result[domain.Object]{Data: domain.Object{"kind": "board"}}, nil
}

func (reader *fakeReader) ListColumns(ctx context.Context, _ domain.ID) (domain.Result[[]domain.Object], error) {
	reader.checkContext(ctx)
	return listResult("columns"), nil
}

func (reader *fakeReader) ListSubcolumns(ctx context.Context, _ domain.ID) (domain.Result[[]domain.Object], error) {
	reader.checkContext(ctx)
	return listResult("subcolumns"), nil
}

func (reader *fakeReader) ListLanes(ctx context.Context, _ domain.ID) (domain.Result[[]domain.Object], error) {
	reader.checkContext(ctx)
	return listResult("lanes"), nil
}

func listResult(kind string) domain.Result[[]domain.Object] {
	return domain.Result[[]domain.Object]{Data: []domain.Object{{"kind": kind}}}
}
