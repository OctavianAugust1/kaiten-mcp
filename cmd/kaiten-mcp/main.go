package main

import (
	"context"
	"fmt"
	"os"

	"github.com/OctavianAugust1/kaiten-mcp/internal/app"
	"github.com/OctavianAugust1/kaiten-mcp/internal/config"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var version = "dev"

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load(".env")
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	application, err := app.New(cfg, version, os.Stderr)
	if err != nil {
		return fmt.Errorf("assemble application: %w", err)
	}
	return application.Server.Run(ctx, &sdk.StdioTransport{})
}
