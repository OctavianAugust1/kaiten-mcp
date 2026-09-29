package app

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/OctavianAugust1/kaiten-mcp/internal/adapter/kaiten"
	"github.com/OctavianAugust1/kaiten-mcp/internal/config"
	mcptransport "github.com/OctavianAugust1/kaiten-mcp/internal/transport/mcp"
	cardusecase "github.com/OctavianAugust1/kaiten-mcp/internal/usecase/card"
	catalogusecase "github.com/OctavianAugust1/kaiten-mcp/internal/usecase/catalog"
	documentusecase "github.com/OctavianAugust1/kaiten-mcp/internal/usecase/document"
	userusecase "github.com/OctavianAugust1/kaiten-mcp/internal/usecase/user"
	workspaceusecase "github.com/OctavianAugust1/kaiten-mcp/internal/usecase/workspace"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Application contains the assembled stdio MCP server and its process logger.
type Application struct {
	Server *sdk.Server
	Logger *slog.Logger
}

// New builds the entire read-only dependency graph with explicit constructors.
func New(cfg config.Config, version string, stderr io.Writer) (*Application, error) {
	logger := NewLogger(cfg.LogLevel(), stderr)
	httpClient := &http.Client{Transport: kaiten.NewReadOnlyTransport(http.DefaultTransport)}
	baseURL := cfg.BaseURL()
	reader, err := kaiten.NewClient(baseURL, cfg.Token(), version, httpClient, cfg.RequestTimeout())
	if err != nil {
		return nil, err
	}
	server, err := mcptransport.NewServer(version, cfg.DefaultPageLimit(), mcptransport.Services{
		Workspace: workspaceusecase.NewService(reader),
		Card:      cardusecase.NewService(reader, cfg.DefaultPageLimit()),
		User:      userusecase.NewService(reader),
		Catalog:   catalogusecase.NewService(reader),
		Document:  documentusecase.NewService(reader, cfg.DefaultPageLimit()),
	})
	if err != nil {
		return nil, err
	}
	return &Application{Server: server, Logger: logger}, nil
}
