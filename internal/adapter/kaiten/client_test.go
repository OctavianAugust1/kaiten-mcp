package kaiten

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

func TestGetJSONBuildsIdentifiedAuthenticatedGETRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Errorf("request method = %q, want GET", request.Method)
		}
		if request.URL.EscapedPath() != "/api/v1/cards" || request.URL.RawQuery != "limit=20&query=urgent" {
			t.Errorf("request URL = %q?%s, want /api/v1/cards?limit=20&query=urgent", request.URL.EscapedPath(), request.URL.RawQuery)
		}
		for name, want := range map[string]string{
			"Authorization":           "Bearer test-token",
			"Accept":                  "application/json",
			"X-Kaiten-Client":         "kaiten-mcp",
			"X-Kaiten-Client-Version": "v0.1.0-test",
		} {
			if got := request.Header.Get(name); got != want {
				t.Errorf("request header %s = %q, want %q", name, got, want)
			}
		}
		writer.Header().Set("X-Total-Count", "1")
		_, _ = writer.Write([]byte(`{"id":1,"title":"A"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server, WithMaxResponseBytes(1024))
	query := url.Values{"query": {"urgent"}, "limit": {"20"}}
	var got domain.Object
	header, err := client.getJSON(context.Background(), "cards", query, &got)
	if err != nil {
		t.Fatalf("getJSON() error = %v", err)
	}
	if got["title"] != "A" || header.Get("X-Total-Count") != "1" {
		t.Errorf("getJSON() object/header = %#v/%q, want preserved payload/header", got, header.Get("X-Total-Count"))
	}
}

func TestGetJSONRetriesTemporaryFailuresWithoutRealSleep(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name       string
		failStatus int
		retryAfter string
		wantWait   time.Duration
	}{
		{name: "request timeout", failStatus: http.StatusRequestTimeout, wantWait: 100 * time.Millisecond},
		{name: "rate limit", failStatus: http.StatusTooManyRequests, retryAfter: "5", wantWait: 5 * time.Second},
		{name: "temporary server failure", failStatus: http.StatusServiceUnavailable, wantWait: 100 * time.Millisecond},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var attempts atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				if attempts.Add(1) == 1 {
					if tt.retryAfter != "" {
						writer.Header().Set("Retry-After", tt.retryAfter)
					}
					writer.WriteHeader(tt.failStatus)
					return
				}
				_, _ = writer.Write([]byte(`{"ok":true}`))
			}))
			defer server.Close()

			var waits []time.Duration
			client := newTestClient(t, server, withRetryPolicy(
				3,
				func(_ context.Context, wait time.Duration) error {
					waits = append(waits, wait)
					return nil
				},
				func() time.Time { return time.Unix(1_000, 0) },
				func(int) time.Duration { return 100 * time.Millisecond },
			))
			var got domain.Object
			if _, err := client.getJSON(context.Background(), "cards", nil, &got); err != nil {
				t.Fatalf("getJSON() error = %v", err)
			}
			if attempts.Load() != 2 || len(waits) != 1 || waits[0] != tt.wantWait {
				t.Errorf("attempts/waits = %d/%v, want 2/[%s]", attempts.Load(), waits, tt.wantWait)
			}
		})
	}
}

func TestGetJSONRetriesNetworkFailureThenSucceeds(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	doer := doerFunc(func(request *http.Request) (*http.Response, error) {
		if attempts.Add(1) == 1 {
			return nil, errors.New("temporary connection failure")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			Request:    request,
		}, nil
	})
	client := newClientWithDoer(t, doer, withRetryPolicy(
		2,
		func(context.Context, time.Duration) error { return nil },
		time.Now,
		func(int) time.Duration { return 0 },
	))
	var got domain.Object
	if _, err := client.getJSON(context.Background(), "cards", nil, &got); err != nil {
		t.Fatalf("getJSON() error = %v", err)
	}
	if attempts.Load() != 2 || got["ok"] != true {
		t.Errorf("attempts/result = %d/%#v, want 2/ok", attempts.Load(), got)
	}
}

func TestGetJSONReturnsTypedErrorAfterRetryExhaustion(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	doer := doerFunc(func(*http.Request) (*http.Response, error) {
		attempts.Add(1)
		return nil, errors.New("network unavailable")
	})
	client := newClientWithDoer(t, doer, withRetryPolicy(
		3,
		func(context.Context, time.Duration) error { return nil },
		time.Now,
		func(int) time.Duration { return 0 },
	))
	var got domain.Object
	_, err := client.getJSON(context.Background(), "cards", nil, &got)
	var upstreamErr *domain.Error
	if !errors.As(err, &upstreamErr) || upstreamErr.Kind != domain.ErrorTemporaryUpstream {
		t.Fatalf("getJSON() error = %v, want temporary-upstream domain error", err)
	}
	if attempts.Load() != 3 {
		t.Errorf("attempts = %d, want 3", attempts.Load())
	}
}

func TestGetJSONReturnsTimeoutAndCancellation(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		ctx      func() context.Context
		do       func(*http.Request) (*http.Response, error)
		wantKind domain.ErrorKind
	}{
		{
			name: "request timeout",
			ctx:  context.Background,
			do: func(*http.Request) (*http.Response, error) {
				return nil, context.DeadlineExceeded
			},
			wantKind: domain.ErrorTimeout,
		},
		{
			name: "caller cancellation",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			do: func(*http.Request) (*http.Response, error) {
				t.Fatal("HTTP transport called after caller cancellation")
				return nil, nil
			},
			wantKind: domain.ErrorCancelled,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := newClientWithDoer(t, doerFunc(tt.do), withRetryPolicy(
				1,
				func(context.Context, time.Duration) error { return nil },
				time.Now,
				func(int) time.Duration { return 0 },
			))
			var got domain.Object
			_, err := client.getJSON(tt.ctx(), "cards", nil, &got)
			var upstreamErr *domain.Error
			if !errors.As(err, &upstreamErr) || upstreamErr.Kind != tt.wantKind {
				t.Fatalf("getJSON() error = %v, want kind %s", err, tt.wantKind)
			}
		})
	}
}

func TestGetJSONRejectsMalformedOrOversizedSuccessBody(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name string
		body string
		max  int64
	}{
		{name: "malformed JSON", body: `{"id":`, max: 1024},
		{name: "oversized body", body: `{"id":123456789}`, max: 8},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				_, _ = writer.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := newTestClient(t, server, WithMaxResponseBytes(tt.max))
			var got domain.Object
			_, err := client.getJSON(context.Background(), "cards", nil, &got)
			var upstreamErr *domain.Error
			if !errors.As(err, &upstreamErr) || upstreamErr.Kind != domain.ErrorMalformedResponse {
				t.Fatalf("getJSON() error = %v, want malformed-response domain error", err)
			}
		})
	}
}

func TestGetJSONCategorizesNonRetryableStatusWithoutLeakingBody(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte("secret upstream diagnostic"))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	var got domain.Object
	_, err := client.getJSON(context.Background(), "cards", nil, &got)
	var upstreamErr *domain.Error
	if !errors.As(err, &upstreamErr) || upstreamErr.Kind != domain.ErrorUpstream || upstreamErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("getJSON() error = %v, want non-retryable upstream domain error", err)
	}
	if attempts.Load() != 1 {
		t.Errorf("attempts = %d, want 1", attempts.Load())
	}
	if strings.Contains(err.Error(), "secret upstream diagnostic") {
		t.Fatalf("error exposed upstream body: %v", err)
	}
}

func newTestClient(t *testing.T, server *httptest.Server, options ...Option) *Client {
	t.Helper()
	baseURL, err := url.Parse(server.URL + "/api/v1/")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	client, err := NewClient(*baseURL, "test-token", "v0.1.0-test", server.Client(), time.Second, options...)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client
}

type doerFunc func(*http.Request) (*http.Response, error)

func (function doerFunc) Do(request *http.Request) (*http.Response, error) {
	return function(request)
}

func newClientWithDoer(t *testing.T, doer httpDoer, options ...Option) *Client {
	t.Helper()
	baseURL, err := url.Parse("https://example.kaiten.ru/api/v1/")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	client, err := NewClient(*baseURL, "test-token", "v0.1.0-test", doer, time.Second, options...)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client
}
