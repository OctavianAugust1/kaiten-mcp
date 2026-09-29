# Development guide

## API source of truth

Use the [official Kaiten API documentation](https://developers.kaiten.ru/) for endpoint paths, query fields, identifiers, and response behavior. The projects under `other-kaiten-mcp/` are examples only. Do not copy an endpoint or parameter from them unless the official documentation confirms it.

The adapter has these invariants:

- `KAITEN_BASE_URL` is an absolute HTTPS API URL such as `https://example.kaiten.ru/api/v1` or `https://example.kaiten.ru/api/latest`.
- Domain reader methods select fixed relative paths. No caller can supply an HTTP method or arbitrary endpoint URL.
- Requests pass through a GET-only transport and carry `Authorization: Bearer ...`, `Accept: application/json`, `X-kaiten-Client: kaiten-mcp`, and `X-kaiten-Client-Version`.
- Query values use `net/url`; identifiers are validated before path construction.
- A list call reads at most one page. It returns continuation metadata and never follows it automatically.
- Response bodies, retries, timeouts, and error text are bounded. Tokens and upstream bodies are not logged or returned to MCP clients.

Relevant official read endpoints are linked from the [Kaiten documentation sitemap](https://developers.kaiten.ru/sitemap.xml), including the pages for [spaces](https://developers.kaiten.ru/spaces/retrieve-list-of-spaces), [cards](https://developers.kaiten.ru/cards/retrieve-card-list), [users](https://developers.kaiten.ru/users/retrieve-list-of-users), and [documents](https://developers.kaiten.ru/documents/retrieve-list-of-documents).
