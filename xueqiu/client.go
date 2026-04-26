package xueqiu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const quoteURL = "https://stock.xueqiu.com/v5/stock/quote.json"

// CookieProvider supplies the Cookie header used to call stock.xueqiu.com.
type CookieProvider interface {
	Cookie(context.Context) (string, error)
}

// Quote is the full upstream quote API response.
//
// Numeric values are decoded as json.Number to avoid precision loss when
// callers inspect large fields such as amount, volume, and timestamps.
type Quote map[string]any

// Client queries Xueqiu realtime quote data.
type Client struct {
	httpClient     *http.Client
	cookieProvider CookieProvider
	cookie         string
}

// NewClient creates a quote client using the provided HTTP client and cookie provider.
//
// Pass nil for httpClient to use http.DefaultClient.
func NewClient(httpClient *http.Client, cookieProvider CookieProvider) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		httpClient:     httpClient,
		cookieProvider: cookieProvider,
	}
}

// ParseSymbols parses a comma-separated symbol list and normalizes symbols to uppercase.
func ParseSymbols(csv string) ([]string, error) {
	parts := strings.Split(csv, ",")
	symbols := make([]string, 0, len(parts))
	for _, part := range parts {
		symbol := strings.ToUpper(strings.TrimSpace(part))
		if symbol == "" {
			continue
		}
		symbols = append(symbols, symbol)
	}
	if len(symbols) == 0 {
		return nil, fmt.Errorf("symbol is required")
	}
	return symbols, nil
}

// Quote queries one symbol and returns the full upstream API response.
func (c *Client) Quote(ctx context.Context, symbol string) (Quote, error) {
	symbols, err := ParseSymbols(symbol)
	if err != nil {
		return nil, err
	}
	if len(symbols) != 1 {
		return nil, fmt.Errorf("expected one symbol, got %d", len(symbols))
	}

	cookie, err := c.getCookie(ctx)
	if err != nil {
		return nil, err
	}

	return c.quoteWithCookie(ctx, symbols[0], cookie)
}

// Quotes queries multiple symbols and returns one full upstream API response per symbol.
func (c *Client) Quotes(ctx context.Context, symbols []string) ([]Quote, error) {
	cookie, err := c.getCookie(ctx)
	if err != nil {
		return nil, err
	}

	quotes := make([]Quote, 0, len(symbols))
	for _, symbol := range symbols {
		parsed, err := ParseSymbols(symbol)
		if err != nil {
			return nil, err
		}
		if len(parsed) != 1 {
			return nil, fmt.Errorf("expected one symbol, got %d", len(parsed))
		}

		quote, err := c.quoteWithCookie(ctx, parsed[0], cookie)
		if err != nil {
			return nil, err
		}
		quotes = append(quotes, quote)
	}
	return quotes, nil
}

func (c *Client) getCookie(ctx context.Context) (string, error) {
	if c.cookie != "" {
		return c.cookie, nil
	}
	if c.cookieProvider == nil {
		return "", fmt.Errorf("cookie provider is required")
	}

	cookie, err := c.cookieProvider.Cookie(ctx)
	if err != nil {
		return "", fmt.Errorf("get xueqiu cookie: %w", err)
	}
	if strings.TrimSpace(cookie) == "" {
		return "", fmt.Errorf("get xueqiu cookie: empty cookie")
	}
	c.cookie = cookie
	return cookie, nil
}

func (c *Client) quoteWithCookie(ctx context.Context, symbol string, cookie string) (Quote, error) {
	reqURL, err := url.Parse(quoteURL)
	if err != nil {
		return nil, err
	}
	params := reqURL.Query()
	params.Set("symbol", symbol)
	params.Set("extend", "detail")
	reqURL.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://xueqiu.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Cookie", cookie)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("xueqiu quote status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	var response Quote
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("decode xueqiu quote response: %w", err)
	}
	if errorCode := quoteErrorCode(response); errorCode != 0 {
		return nil, fmt.Errorf("xueqiu quote error %d: %s", errorCode, quoteErrorDescription(response))
	}

	return response, nil
}

func quoteErrorCode(response Quote) int {
	value, ok := response["error_code"]
	if !ok {
		return 0
	}

	switch typed := value.(type) {
	case json.Number:
		n, _ := typed.Int64()
		return int(n)
	case float64:
		return int(typed)
	case int:
		return typed
	default:
		return 0
	}
}

func quoteErrorDescription(response Quote) string {
	value, ok := response["error_description"]
	if !ok {
		return ""
	}
	if typed, ok := value.(string); ok {
		return typed
	}
	return fmt.Sprint(value)
}
