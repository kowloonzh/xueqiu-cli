package xueqiu_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/kowloonzh/xueqiu-cli/xueqiu"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type fakeCookieProvider struct {
	cookie string
}

func (p fakeCookieProvider) Cookie(context.Context) (string, error) {
	return p.cookie, nil
}

func TestPublicClientCanBeUsedByExternalGoPackages(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Query().Get("symbol") != "SH513310" {
			t.Fatalf("symbol query = %q", req.URL.Query().Get("symbol"))
		}
		if req.Header.Get("Cookie") != "xq_a_token=abc" {
			t.Fatalf("Cookie header = %q", req.Header.Get("Cookie"))
		}

		body := `{
			"data": {
				"quote": {
					"symbol": "SH513310",
					"name": "中韩半导体ETF华泰柏瑞",
					"premium_rate": 6.54
				}
			},
			"error_code": 0,
			"error_description": ""
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})}

	client := xueqiu.NewClient(httpClient, fakeCookieProvider{cookie: "xq_a_token=abc"})
	quote, err := client.Quote(context.Background(), "sh513310")
	if err != nil {
		t.Fatalf("Quote returned error: %v", err)
	}

	data := quote["data"].(map[string]any)
	quoteData := data["quote"].(map[string]any)
	if quoteData["symbol"] != "SH513310" {
		t.Fatalf("symbol = %v", quoteData["symbol"])
	}
	premiumRate, ok := quoteData["premium_rate"].(json.Number)
	if !ok {
		t.Fatalf("premium_rate type = %T", quoteData["premium_rate"])
	}
	if premiumRate.String() != "6.54" {
		t.Fatalf("premium_rate = %v", premiumRate)
	}
}
