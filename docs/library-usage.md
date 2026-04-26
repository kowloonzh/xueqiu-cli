# Go Library Usage

`xueqiu-cli` can also be used as a Go dependency through the public package:

```go
import "github.com/kowloonzh/xueqiu-cli/xueqiu"
```

The CLI entrypoint remains in the module root. Library consumers should import the `xueqiu` subpackage, not the module root package.

## Install

```bash
go get github.com/kowloonzh/xueqiu-cli
```

## Query One Symbol

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kowloonzh/xueqiu-cli/xueqiu"
)

func main() {
	provider := xueqiu.NewChromeCookieProvider(xueqiu.ChromeConfig{
		Timeout: 30 * time.Second,
	})
	client := xueqiu.NewClient(http.DefaultClient, provider)

	quote, err := client.Quote(context.Background(), "SH513310")
	if err != nil {
		panic(err)
	}

	b, _ := json.MarshalIndent(quote, "", "  ")
	fmt.Println(string(b))
}
```

## Query Multiple Symbols

```go
symbols, err := xueqiu.ParseSymbols("SH513310,SH000001")
if err != nil {
	panic(err)
}

quotes, err := client.Quotes(context.Background(), symbols)
if err != nil {
	panic(err)
}
```

## Chrome Remote Mode

If you already have a Chrome DevTools endpoint:

```go
provider := xueqiu.NewChromeCookieProvider(xueqiu.ChromeConfig{
	RemoteURL: "http://127.0.0.1:9222",
	Timeout:   30 * time.Second,
})
client := xueqiu.NewClient(http.DefaultClient, provider)
```

The remote endpoint must expose `/json/version`. If `curl http://127.0.0.1:9222/json/version` returns 404 or EOF, it is not a usable Chrome DevTools endpoint for this library.

## Response Shape

`Client.Quote` returns `xueqiu.Quote`, which is:

```go
type Quote map[string]any
```

The map contains the full upstream Xueqiu response, including:

```text
data.market
data.others
data.quote
data.tags
error_code
error_description
```

Numeric fields are decoded as `json.Number` so large values such as `amount`, `volume`, and timestamps are not rounded by `float64` conversion.

ETF fields such as `premium_rate`, `unit_nav`, and `acc_unit_nav` are available under:

```text
data.quote.premium_rate
data.quote.unit_nav
data.quote.acc_unit_nav
```

