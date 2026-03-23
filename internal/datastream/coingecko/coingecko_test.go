package coingecko

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func withMockTransport(t *testing.T, fn roundTripFunc) {
	t.Helper()
	original := http.DefaultTransport
	http.DefaultTransport = fn
	t.Cleanup(func() {
		http.DefaultTransport = original
	})
}

func TestFetchPriceWithMock(t *testing.T) {
	feed := New()
	ctx := context.Background()

	withMockTransport(t, func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "api.coingecko.com" {
			t.Fatalf("unexpected host: %s", req.URL.Host)
		}

		if got := req.URL.Query().Get("ids"); got != "bitcoin" {
			t.Fatalf("unexpected ids query param: %s", got)
		}

		if got := req.URL.Query().Get("vs_currencies"); got != "usd" {
			t.Fatalf("unexpected vs_currencies query param: %s", got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"bitcoin":{"usd":123.45}}`)),
			Header:     make(http.Header),
		}, nil
	})

	price, err := feed.FetchPrice(ctx, "bitcoin", "0x")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if price == nil {
		t.Fatal("expected price, got nil")
	}

	if price.Value != 123.45 {
		t.Fatalf("expected value 123.45, got: %f", price.Value)
	}

	if price.Source != "coingecko" {
		t.Fatalf("expected source coingecko, got: %s", price.Source)
	}

	if price.Timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
}

func TestFetchPriceRealHTTPParsesStructure(t *testing.T) {
	feed := New()
	ctx := context.Background()

	price, err := feed.FetchPrice(ctx, "zarp-stablecoin", "0x")
	if err != nil {
		t.Fatalf("expected no error from real HTTP call, got: %v", err)
	}

	if price == nil {
		t.Fatal("expected price, got nil")
	}

	if price.Source != "coingecko" {
		t.Fatalf("expected source coingecko, got: %s", price.Source)
	}

	if price.Timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}

	if price.Value <= 0 {
		t.Fatalf("expected positive parsed USD value, got: %f", price.Value)
	}

	p := price.ToUnified()

	t.Logf("price %f", p.Value)
}
