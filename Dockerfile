FROM golang:1.27-bookworm AS development

RUN apt-get update && apt-get install -y --no-install-recommends \
	build-essential \
	ca-certificates \
	curl \
	git \
	gnupg \
	jq \
	&& mkdir -p /etc/apt/keyrings \
	&& curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
	&& apt-get install -y --no-install-recommends nodejs \
	&& rm -rf /var/lib/apt/lists/*

ENV NPM_CONFIG_PREFIX=/usr/local
ENV GOPATH=/go
ENV PATH=/usr/local/bin:/go/bin:/usr/local/go/bin:/usr/bin:/bin

RUN go install golang.org/x/tools/gopls@latest \
	&& go install github.com/go-delve/delve/cmd/dlv@latest \
	&& go install github.com/air-verse/air@latest \
	&& go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

WORKDIR /workspace
CMD ["bash"]

FROM golang:1.27-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/kaiten-mcp ./cmd/kaiten-mcp

FROM gcr.io/distroless/static-debian12:nonroot AS runtime

COPY --from=build --chown=nonroot:nonroot /out/kaiten-mcp /usr/local/bin/kaiten-mcp
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/kaiten-mcp"]
