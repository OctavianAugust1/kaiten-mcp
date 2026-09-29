package kaiten

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

const defaultMaxResponseBytes int64 = 8 << 20

type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Client performs authenticated reads against fixed Kaiten endpoint paths.
type Client struct {
	baseURL          url.URL
	token            string
	clientVersion    string
	httpClient       httpDoer
	requestTimeout   time.Duration
	maxResponseBytes int64
	maxAttempts      int
	sleep            func(context.Context, time.Duration) error
	now              func() time.Time
	backoff          func(int) time.Duration
}

func withRetryPolicy(
	maxAttempts int,
	sleep func(context.Context, time.Duration) error,
	now func() time.Time,
	backoff func(int) time.Duration,
) Option {
	return func(client *Client) error {
		if maxAttempts < 1 || sleep == nil || now == nil || backoff == nil {
			return errors.New("retry policy requires positive attempts and non-nil hooks")
		}
		client.maxAttempts = maxAttempts
		client.sleep = sleep
		client.now = now
		client.backoff = backoff
		return nil
	}
}

// Option adjusts bounded client behavior without exposing HTTP method choice.
type Option func(*Client) error

// WithMaxResponseBytes bounds the successful JSON body read before decoding.
func WithMaxResponseBytes(limit int64) Option {
	return func(client *Client) error {
		if limit < 1 {
			return errors.New("response byte limit must be positive")
		}
		client.maxResponseBytes = limit
		return nil
	}
}

// NewClient builds a GET-only Kaiten adapter.
func NewClient(
	baseURL url.URL,
	token string,
	clientVersion string,
	httpClient httpDoer,
	requestTimeout time.Duration,
	options ...Option,
) (*Client, error) {
	if baseURL.Scheme != "https" || baseURL.Host == "" || baseURL.User != nil || baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return nil, errors.New("base URL must be an absolute HTTPS URL without user info, query, or fragment")
	}
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("token is required")
	}
	if strings.TrimSpace(clientVersion) == "" {
		return nil, errors.New("client version is required")
	}
	if httpClient == nil {
		return nil, errors.New("HTTP client is required")
	}
	if requestTimeout <= 0 {
		return nil, errors.New("request timeout must be positive")
	}
	baseURL.Path = strings.TrimRight(baseURL.Path, "/") + "/"
	client := &Client{
		baseURL:          baseURL,
		token:            strings.TrimSpace(token),
		clientVersion:    strings.TrimSpace(clientVersion),
		httpClient:       httpClient,
		requestTimeout:   requestTimeout,
		maxResponseBytes: defaultMaxResponseBytes,
		maxAttempts:      3,
		sleep:            sleepContext,
		now:              time.Now,
		backoff:          retryBackoff,
	}
	for _, option := range options {
		if err := option(client); err != nil {
			return nil, fmt.Errorf("configure Kaiten client: %w", err)
		}
	}
	return client, nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, query url.Values, destination any) (http.Header, error) {
	requestURL, err := c.resolveEndpoint(endpoint, query)
	if err != nil {
		return nil, &domain.Error{Kind: domain.ErrorInvalidArgument, Operation: "build Kaiten GET", Cause: err}
	}
	if err := ctx.Err(); err != nil {
		return nil, classifyRequestError(ctx, err)
	}

	var lastErr error
	for attempt := 1; attempt <= c.maxAttempts; attempt++ {
		header, attemptErr := c.getJSONAttempt(ctx, requestURL, destination)
		if attemptErr == nil {
			return header, nil
		}
		lastErr = attemptErr
		if attempt == c.maxAttempts || !retryable(attemptErr) {
			break
		}
		wait := c.backoff(attempt)
		var upstreamErr *domain.Error
		if errors.As(attemptErr, &upstreamErr) && upstreamErr.RetryAfter > 0 {
			wait = upstreamErr.RetryAfter
		}
		if err := c.sleep(ctx, wait); err != nil {
			return nil, classifyRequestError(ctx, err)
		}
	}
	return nil, lastErr
}

func (c *Client) getJSONAttempt(ctx context.Context, requestURL url.URL, destination any) (http.Header, error) {
	attemptCtx, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(attemptCtx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, &domain.Error{Kind: domain.ErrorInvalidArgument, Operation: "build Kaiten GET", Cause: err}
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-kaiten-Client", "kaiten-mcp")
	request.Header.Set("X-kaiten-Client-Version", c.clientVersion)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, classifyRequestError(attemptCtx, err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, c.maxResponseBytes))
		return nil, statusError(response.StatusCode, parseRetryAfter(response.Header.Get("Retry-After"), c.now()))
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, c.maxResponseBytes+1))
	if err != nil {
		return nil, &domain.Error{Kind: domain.ErrorMalformedResponse, Operation: "read Kaiten response", Cause: err}
	}
	if int64(len(body)) > c.maxResponseBytes {
		return nil, &domain.Error{Kind: domain.ErrorMalformedResponse, Operation: "read Kaiten response"}
	}
	if err := decodeJSON(body, destination); err != nil {
		return nil, err
	}
	return response.Header.Clone(), nil
}

func decodeJSON(body []byte, destination any) error {
	if err := json.Unmarshal(body, destination); err != nil {
		return &domain.Error{Kind: domain.ErrorMalformedResponse, Operation: "decode Kaiten response", Cause: err}
	}
	return nil
}

func (c *Client) resolveEndpoint(endpoint string, query url.Values) (url.URL, error) {
	if endpoint == "" || strings.HasPrefix(endpoint, "/") || strings.ContainsAny(endpoint, "?#\\") || path.Clean(endpoint) != endpoint {
		return url.URL{}, errors.New("endpoint must be a clean relative path")
	}
	relative := url.URL{Path: endpoint, RawQuery: query.Encode()}
	return *c.baseURL.ResolveReference(&relative), nil
}

func classifyRequestError(ctx context.Context, err error) error {
	if errors.Is(ctx.Err(), context.Canceled) || errors.Is(err, context.Canceled) {
		return &domain.Error{Kind: domain.ErrorCancelled, Operation: "perform Kaiten GET", Cause: err}
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		return &domain.Error{Kind: domain.ErrorTimeout, Operation: "perform Kaiten GET", Cause: err}
	}
	return &domain.Error{Kind: domain.ErrorTemporaryUpstream, Operation: "perform Kaiten GET", Cause: err}
}

func statusError(statusCode int, retryAfter time.Duration) error {
	kind := domain.ErrorUpstream
	switch statusCode {
	case http.StatusUnauthorized:
		kind = domain.ErrorUnauthorized
	case http.StatusForbidden:
		kind = domain.ErrorForbidden
	case http.StatusNotFound:
		kind = domain.ErrorNotFound
	case http.StatusRequestTimeout:
		kind = domain.ErrorTemporaryUpstream
	case http.StatusTooManyRequests:
		kind = domain.ErrorRateLimited
	default:
		if statusCode >= http.StatusInternalServerError {
			kind = domain.ErrorTemporaryUpstream
		}
	}
	return &domain.Error{Kind: kind, Operation: "Kaiten GET failed", StatusCode: statusCode, RetryAfter: retryAfter}
}

func retryable(err error) bool {
	var upstreamErr *domain.Error
	if !errors.As(err, &upstreamErr) {
		return false
	}
	switch upstreamErr.Kind {
	case domain.ErrorRateLimited, domain.ErrorTimeout, domain.ErrorTemporaryUpstream:
		return true
	default:
		return false
	}
}

func parseRetryAfter(value string, now time.Time) time.Duration {
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	when, err := http.ParseTime(value)
	if err != nil || !when.After(now) {
		return 0
	}
	return when.Sub(now)
}

func sleepContext(ctx context.Context, wait time.Duration) error {
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func retryBackoff(attempt int) time.Duration {
	base := 100 * time.Millisecond
	for index := 1; index < attempt && base < 2*time.Second; index++ {
		base *= 2
	}
	if base > 2*time.Second {
		base = 2 * time.Second
	}
	half := base / 2
	// Pseudorandom jitter prevents synchronized retries; it is not used for a
	// security-sensitive value.
	return half + time.Duration(rand.Int64N(int64(half)+1)) //nolint:gosec // retry jitter is non-cryptographic.
}
