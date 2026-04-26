package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestGetRawClassifies429AsRateLimit(t *testing.T) {
	client := &Client{
		host: "https://api.github.com",
		httpCli: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Body:       io.NopCloser(strings.NewReader("rate limited")),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	_, err := client.GetRaw("/raw")
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("GetRaw() error = %T %v, want *APIError", err, err)
	}
	if apiErr.Kind != "rate_limit" || apiErr.Status != http.StatusTooManyRequests || apiErr.ExitCode != 4 {
		t.Fatalf("GetRaw() APIError = %#v, want rate_limit status 429 exit 4", apiErr)
	}
}
