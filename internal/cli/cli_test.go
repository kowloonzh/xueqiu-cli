package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/kowloonzh/xueqiu-cli/xueqiu"
)

type fakeQuoteService struct {
	symbols []string
}

func (s *fakeQuoteService) Quote(ctx context.Context, symbol string) (xueqiu.Quote, error) {
	s.symbols = append(s.symbols, symbol)
	return xueqiu.Quote{
		"data": map[string]any{
			"quote": map[string]any{
				"symbol":       symbol,
				"name":         "测试",
				"current":      12.34,
				"premium_rate": 6.54,
			},
		},
		"error_code": float64(0),
	}, nil
}

func (s *fakeQuoteService) Quotes(ctx context.Context, symbols []string) ([]xueqiu.Quote, error) {
	s.symbols = append(s.symbols, symbols...)
	quotes := make([]xueqiu.Quote, 0, len(symbols))
	for _, symbol := range symbols {
		quotes = append(quotes, xueqiu.Quote{
			"data": map[string]any{
				"quote": map[string]any{
					"symbol":       symbol,
					"name":         "测试",
					"current":      12.34,
					"premium_rate": 6.54,
				},
			},
			"error_code": float64(0),
		})
	}
	return quotes, nil
}

func TestStockCommandPrintsSingleQuoteJSON(t *testing.T) {
	var out bytes.Buffer
	service := &fakeQuoteService{}
	cmd := NewRootCommandWithService(service, &out)
	cmd.SetArgs([]string{"stock", "sh000001"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out.String())
	}
	data := got["data"].(map[string]any)
	quote := data["quote"].(map[string]any)
	if quote["symbol"] != "SH000001" {
		t.Fatalf("symbol = %q", quote["symbol"])
	}
	if quote["premium_rate"] != 6.54 {
		t.Fatalf("premium_rate = %v", quote["premium_rate"])
	}
	if len(service.symbols) != 1 || service.symbols[0] != "SH000001" {
		t.Fatalf("service symbols = %#v", service.symbols)
	}
}

func TestStocksCommandPrintsQuoteArrayJSON(t *testing.T) {
	var out bytes.Buffer
	service := &fakeQuoteService{}
	cmd := NewRootCommandWithService(service, &out)
	cmd.SetArgs([]string{"stocks", "sh000001,csih30269"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	var got []map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out.String())
	}
	if len(got) != 2 {
		t.Fatalf("len(output) = %d", len(got))
	}
	firstQuote := got[0]["data"].(map[string]any)["quote"].(map[string]any)
	secondQuote := got[1]["data"].(map[string]any)["quote"].(map[string]any)
	if firstQuote["symbol"] != "SH000001" || secondQuote["symbol"] != "CSIH30269" {
		t.Fatalf("symbols = %#v", got)
	}
}

func TestChromeRemoteURLSelectsRemoteMode(t *testing.T) {
	var out bytes.Buffer
	var got Config
	cmd := newRootCommand(&out, func(config Config) QuoteService {
		got = config
		return &fakeQuoteService{}
	})
	cmd.SetArgs([]string{"--chrome-remote-url=ws://127.0.0.1:9222/devtools/browser/test", "stock", "sh000001"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if got.ChromeRemoteURL != "ws://127.0.0.1:9222/devtools/browser/test" {
		t.Fatalf("ChromeRemoteURL = %q", got.ChromeRemoteURL)
	}
}
