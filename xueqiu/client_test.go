package xueqiu

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type fakeCookieProvider struct {
	cookie string
	calls  int
}

func (p *fakeCookieProvider) Cookie(context.Context) (string, error) {
	p.calls++
	return p.cookie, nil
}

func TestParseSymbolsTrimsUppercasesAndRejectsEmpty(t *testing.T) {
	got, err := ParseSymbols(" sh000001,csih30269 ")
	if err != nil {
		t.Fatalf("ParseSymbols returned error: %v", err)
	}

	want := []string{"SH000001", "CSIH30269"}
	if len(got) != len(want) {
		t.Fatalf("len(ParseSymbols) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ParseSymbols()[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	if _, err := ParseSymbols(" , "); err == nil {
		t.Fatal("ParseSymbols accepted an empty symbol list")
	}
}

func TestClientQuoteSendsCookieAndReturnsFullAPIResponse(t *testing.T) {
	provider := &fakeCookieProvider{cookie: "xq_a_token=abc"}
	httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got := req.URL.String(); got != "https://stock.xueqiu.com/v5/stock/quote.json?extend=detail&symbol=SH000001" {
			t.Fatalf("request URL = %q", got)
		}
		if got := req.Header.Get("Cookie"); got != "xq_a_token=abc" {
			t.Fatalf("Cookie header = %q", got)
		}
		if got := req.Header.Get("Referer"); got != "https://xueqiu.com" {
			t.Fatalf("Referer header = %q", got)
		}

		body := `{
			"data": {
				"market": {"status": "交易中"},
				"quote": {
					"symbol": "SH000001",
					"code": "000001",
					"name": "上证指数",
					"current": 3012.34,
					"percent": 0.56,
					"chg": 16.78,
					"amount": 12340000000,
					"volume": 987654,
					"open": 3000.1,
					"high": 3020.2,
					"low": 2990.3,
					"last_close": 2995.56,
					"turnover_rate": 1.23,
					"timestamp": 1710000000000,
					"currency": "CNY",
					"exchange": "SH",
					"premium_rate": 6.54,
					"unit_nav": 3.688,
					"extra_field_from_api": "kept"
				}
			},
			"error_code": 0
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})}

	client := NewClient(httpClient, provider)
	quote, err := client.Quote(context.Background(), "sh000001")
	if err != nil {
		t.Fatalf("Quote returned error: %v", err)
	}

	data := quote["data"].(map[string]any)
	quoteData := data["quote"].(map[string]any)
	market := data["market"].(map[string]any)

	if quoteData["symbol"] != "SH000001" {
		t.Fatalf("symbol = %q", quoteData["symbol"])
	}
	if quoteData["name"] != "上证指数" {
		t.Fatalf("name = %q", quoteData["name"])
	}
	if quoteData["current"].(json.Number).String() != "3012.34" {
		t.Fatalf("current = %v", quoteData["current"])
	}
	if quoteData["premium_rate"].(json.Number).String() != "6.54" {
		t.Fatalf("premium_rate = %v", quoteData["premium_rate"])
	}
	if quoteData["extra_field_from_api"] != "kept" {
		t.Fatalf("extra_field_from_api = %v", quoteData["extra_field_from_api"])
	}
	if market["status"] != "交易中" {
		t.Fatalf("market status = %q", market["status"])
	}
	if provider.calls != 1 {
		t.Fatalf("cookie provider calls = %d, want 1", provider.calls)
	}
}
