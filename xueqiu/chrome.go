package xueqiu

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
)

// ChromeConfig controls how Chrome is used to obtain a xueqiu.com cookie.
type ChromeConfig struct {
	// RemoteURL is a Chrome DevTools endpoint such as http://127.0.0.1:9222.
	// Leave it empty to launch a local Chrome instance.
	RemoteURL string
	// Timeout bounds the cookie acquisition flow. Defaults to 30 seconds.
	Timeout time.Duration
}

// ChromeCookieProvider obtains and caches a xueqiu.com cookie through Chrome.
type ChromeCookieProvider struct {
	config ChromeConfig
	mu     sync.Mutex
	cookie string
}

// NewChromeCookieProvider creates a cookie provider for local or remote Chrome mode.
func NewChromeCookieProvider(config ChromeConfig) *ChromeCookieProvider {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	return &ChromeCookieProvider{config: config}
}

// Cookie returns a cached xueqiu.com cookie, obtaining one through Chrome if needed.
func (p *ChromeCookieProvider) Cookie(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cookie != "" {
		return p.cookie, nil
	}

	ctx, cancel := context.WithTimeout(ctx, p.config.Timeout)
	defer cancel()

	var allocCtx context.Context
	var allocCancel context.CancelFunc
	if p.config.RemoteURL != "" {
		allocCtx, allocCancel = chromedp.NewRemoteAllocator(ctx, p.config.RemoteURL)
	} else {
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", false),
			chromedp.Flag("enable-automation", false),
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
		)
		allocCtx, allocCancel = chromedp.NewExecAllocator(ctx, opts...)
	}
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	var cookies []*network.Cookie
	err := chromedp.Run(browserCtx,
		chromedp.Navigate("https://xueqiu.com"),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			allCookies, err := storage.GetCookies().Do(ctx)
			if err != nil {
				return err
			}
			for _, cookie := range allCookies {
				if strings.Contains(cookie.Domain, "xueqiu.com") {
					cookies = append(cookies, cookie)
				}
			}
			return nil
		}),
	)
	if err != nil {
		return "", fmt.Errorf("run chrome: %w", err)
	}

	values := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		values = append(values, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}
	if len(values) == 0 {
		return "", fmt.Errorf("no valid cookie found for xueqiu.com")
	}

	p.cookie = strings.Join(values, "; ")
	return p.cookie, nil
}
