package kaiten

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

func TestReadOnlyTransportRejectsNonGETBeforeNetwork(t *testing.T) {
	t.Parallel()

	called := false
	transport := NewReadOnlyTransport(roundTripperFunc(func(*http.Request) (*http.Response, error) {
		called = true
		return nil, errors.New("network should not be reached")
	}))
	request, err := http.NewRequest(http.MethodPost, "https://example.kaiten.ru/api/v1/cards", bytes.NewReader([]byte(`{}`)))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	_, err = transport.RoundTrip(request)
	if err == nil || !errors.Is(err, ErrNonGETRequest) {
		t.Fatalf("RoundTrip() error = %v, want ErrNonGETRequest", err)
	}
	if called {
		t.Fatal("underlying transport was called for POST")
	}
}

func TestReadOnlyTransportPassesGETUnchanged(t *testing.T) {
	t.Parallel()

	transport := NewReadOnlyTransport(roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodGet {
			t.Fatalf("method = %q, want GET", request.Method)
		}
		return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody, Header: make(http.Header)}, nil
	}))
	request, err := http.NewRequest(http.MethodGet, "https://example.kaiten.ru/api/v1/cards", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	response, err := transport.RoundTrip(request)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	if response.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want %d", response.StatusCode, http.StatusNoContent)
	}
}

func FuzzResolveEndpointEncodesQueryAsData(f *testing.F) {
	for _, seed := range []string{"urgent", "a&admin=true", "../cards?method=DELETE", "кириллица", ""} {
		f.Add(seed)
	}
	baseURL, err := url.Parse("https://example.kaiten.ru/api/v1/")
	if err != nil {
		f.Fatalf("Parse() error = %v", err)
	}
	client := Client{baseURL: *baseURL}

	f.Fuzz(func(t *testing.T, value string) {
		got, err := client.resolveEndpoint("cards", url.Values{"query": {value}})
		if err != nil {
			t.Fatalf("resolveEndpoint() error = %v", err)
		}
		if got.Path != "/api/v1/cards" || got.Query().Get("query") != value {
			t.Fatalf("resolved URL = %q, want fixed path and original query value", got.String())
		}
	})
}

func FuzzDecodeJSONRejectsMalformedInput(f *testing.F) {
	for _, seed := range []string{`{"id":1}`, `{"id":`, `null`, "", "not-json"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		var got domain.Object
		err := decodeJSON([]byte(input), &got)
		if !json.Valid([]byte(input)) {
			var upstreamErr *domain.Error
			if !errors.As(err, &upstreamErr) || upstreamErr.Kind != domain.ErrorMalformedResponse {
				t.Fatalf("decodeJSON(%q) error = %v, want malformed-response error", input, err)
			}
		}
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (function roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
