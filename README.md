# xueqiu-cli

`xueqiu-cli` is a small Go CLI and library for querying Xueqiu realtime quote data.

It supports:

- CLI queries for one symbol or multiple symbols
- full upstream JSON output from Xueqiu
- local Chrome mode for cookie acquisition
- Chrome remote mode for server/headless-browser environments
- Go library usage through `github.com/kowloonzh/xueqiu-cli/xueqiu`

## CLI

### Download Binary

Prebuilt binaries are published from GitHub tags:

```text
https://github.com/kowloonzh/xueqiu-cli/releases
```

Choose the archive for your OS and CPU:

```text
xueqiu-cli_1.0.0_darwin_amd64.tar.gz
xueqiu-cli_1.0.0_darwin_arm64.tar.gz
xueqiu-cli_1.0.0_linux_amd64.tar.gz
xueqiu-cli_1.0.0_linux_arm64.tar.gz
xueqiu-cli_1.0.0_windows_amd64.tar.gz
xueqiu-cli_1.0.0_windows_arm64.tar.gz
```

Linux amd64 example:

```bash
curl -L -o xueqiu-cli.tar.gz \
  https://github.com/kowloonzh/xueqiu-cli/releases/download/1.0.0/xueqiu-cli_1.0.0_linux_amd64.tar.gz

tar -xzf xueqiu-cli.tar.gz
sudo install -m 0755 xueqiu-cli_1.0.0_linux_amd64/xueqiu-cli /usr/local/bin/xueqiu-cli
```

macOS Apple Silicon example:

```bash
curl -L -o xueqiu-cli.tar.gz \
  https://github.com/kowloonzh/xueqiu-cli/releases/download/1.0.0/xueqiu-cli_1.0.0_darwin_arm64.tar.gz

tar -xzf xueqiu-cli.tar.gz
sudo install -m 0755 xueqiu-cli_1.0.0_darwin_arm64/xueqiu-cli /usr/local/bin/xueqiu-cli
```

Each release also includes `checksums.txt` for SHA256 verification.

### Build From Source

```bash
git clone git@github.com:kowloonzh/xueqiu-cli.git
cd xueqiu-cli

go test ./...
go build -o xueqiu-cli .
```

Optional install:

```bash
sudo install -m 0755 xueqiu-cli /usr/local/bin/xueqiu-cli
```

### Commands

Print version:

```bash
xueqiu-cli version
```

Query one symbol:

```bash
xueqiu-cli stock SH513310
```

Query multiple symbols:

```bash
xueqiu-cli stocks SH513310,SH000001
```

Output is JSON. `stock` returns one full Xueqiu API response object. `stocks` returns an array of full response objects.

Common ETF fields are under `data.quote`, for example:

```text
data.quote.current
data.quote.percent
data.quote.premium_rate
data.quote.unit_nav
data.quote.acc_unit_nav
data.quote.market_capital
```

See [docs/xueqiu-quote-api.md](docs/xueqiu-quote-api.md) for a field mapping between Xueqiu page labels and JSON fields.

### Chrome Modes

The Xueqiu quote API requires a valid Xueqiu cookie. The CLI gets that cookie through Chrome.

Default mode starts a local Chrome instance:

```bash
xueqiu-cli stock SH513310
```

Remote mode connects to an existing Chrome DevTools endpoint:

```bash
xueqiu-cli --chrome-remote-url=http://127.0.0.1:9222 stock SH513310
```

This is useful on servers where you want to run a headless browser separately.

Start a headless Chrome DevTools endpoint with Docker:

```bash
docker run -d -p 9222:9222 --rm --name headless-shell chromedp/headless-shell
```

Check that the endpoint is usable:

```bash
curl http://127.0.0.1:9222/json/version
```

Then run:

```bash
xueqiu-cli --chrome-remote-url=http://127.0.0.1:9222 stock SH513310
```

Stop the headless browser:

```bash
docker stop headless-shell
```

## Go Library

Install:

```bash
go get github.com/kowloonzh/xueqiu-cli
```

Import the public subpackage:

```go
import "github.com/kowloonzh/xueqiu-cli/xueqiu"
```

### Query One Symbol

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

### Query Multiple Symbols

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

### Library Chrome Remote Mode

```go
provider := xueqiu.NewChromeCookieProvider(xueqiu.ChromeConfig{
	RemoteURL: "http://127.0.0.1:9222",
	Timeout:   30 * time.Second,
})
client := xueqiu.NewClient(http.DefaultClient, provider)
```

`Client.Quote` returns `xueqiu.Quote`, which is:

```go
type Quote map[string]any
```

The map contains the full upstream Xueqiu response:

```text
data.market
data.others
data.quote
data.tags
error_code
error_description
```

Numeric fields are decoded as `json.Number` to avoid precision loss for large fields such as `amount`, `volume`, and timestamps.

More library examples are in [docs/library-usage.md](docs/library-usage.md).

## Release

Pushing a tag builds and publishes release binaries automatically:

```bash
git tag 1.0.0
git push origin 1.0.0
```

The workflow is defined in [.github/workflows/release.yml](.github/workflows/release.yml).

