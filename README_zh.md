# xueqiu-cli

`xueqiu-cli` 是一个用于查询雪球实时行情数据的 Go 命令行工具，同时也可以作为 Go 依赖库使用。

支持能力：

- 查询单个 symbol 或多个 symbols
- 输出雪球上游接口的完整 JSON
- 通过本地 Chrome 获取雪球 cookie
- 支持 Chrome remote 模式，适合服务器无头浏览器场景
- 可作为 Go 库使用：`github.com/kowloonzh/xueqiu-cli/xueqiu`

## CLI 使用

### 下载二进制

预编译二进制会在 GitHub tag 发布时自动生成：

```text
https://github.com/kowloonzh/xueqiu-cli/releases
```

根据操作系统和 CPU 架构选择对应文件：

```text
xueqiu-cli_1.0.0_darwin_amd64.tar.gz
xueqiu-cli_1.0.0_darwin_arm64.tar.gz
xueqiu-cli_1.0.0_linux_amd64.tar.gz
xueqiu-cli_1.0.0_linux_arm64.tar.gz
xueqiu-cli_1.0.0_windows_amd64.tar.gz
xueqiu-cli_1.0.0_windows_arm64.tar.gz
```

Linux amd64 示例：

```bash
curl -L -o xueqiu-cli.tar.gz \
  https://github.com/kowloonzh/xueqiu-cli/releases/download/1.0.0/xueqiu-cli_1.0.0_linux_amd64.tar.gz

tar -xzf xueqiu-cli.tar.gz
sudo install -m 0755 xueqiu-cli_1.0.0_linux_amd64/xueqiu-cli /usr/local/bin/xueqiu-cli
```

macOS Apple Silicon 示例：

```bash
curl -L -o xueqiu-cli.tar.gz \
  https://github.com/kowloonzh/xueqiu-cli/releases/download/1.0.0/xueqiu-cli_1.0.0_darwin_arm64.tar.gz

tar -xzf xueqiu-cli.tar.gz
sudo install -m 0755 xueqiu-cli_1.0.0_darwin_arm64/xueqiu-cli /usr/local/bin/xueqiu-cli
```

每个 Release 也会包含 `checksums.txt`，可用于校验 SHA256。

### 从源码构建

```bash
git clone git@github.com:kowloonzh/xueqiu-cli.git
cd xueqiu-cli

go test ./...
go build -o xueqiu-cli .
```

可选安装：

```bash
sudo install -m 0755 xueqiu-cli /usr/local/bin/xueqiu-cli
```

### 命令说明

查看版本：

```bash
xueqiu-cli version
```

查询单个 symbol：

```bash
xueqiu-cli stock SH513310
```

查询多个 symbols：

```bash
xueqiu-cli stocks SH513310,SH000001
```

输出格式是 JSON。`stock` 返回一个完整雪球接口响应对象，`stocks` 返回完整响应对象数组。

ETF 常用字段在 `data.quote` 下，例如：

```text
data.quote.current
data.quote.percent
data.quote.premium_rate
data.quote.unit_nav
data.quote.acc_unit_nav
data.quote.market_capital
```

页面展示字段和 JSON 字段的对应关系见 [docs/xueqiu-quote-api.md](docs/xueqiu-quote-api.md)。

### Chrome 模式

雪球行情接口需要有效的雪球 cookie。CLI 会通过 Chrome 获取 cookie 后再请求行情接口。

默认模式会启动本地 Chrome：

```bash
xueqiu-cli stock SH513310
```

remote 模式会连接已有的 Chrome DevTools endpoint：

```bash
xueqiu-cli --chrome-remote-url=http://127.0.0.1:9222 stock SH513310
```

这个模式适合在服务器上运行无头浏览器。可以用 Docker 启动 headless Chrome：

```bash
docker run -d -p 9222:9222 --rm --name headless-shell chromedp/headless-shell
```

确认 endpoint 可用：

```bash
curl http://127.0.0.1:9222/json/version
```

然后运行：

```bash
xueqiu-cli --chrome-remote-url=http://127.0.0.1:9222 stock SH513310
```

停止无头浏览器：

```bash
docker stop headless-shell
```

## Go 库使用

安装依赖：

```bash
go get github.com/kowloonzh/xueqiu-cli
```

引入公共子包：

```go
import "github.com/kowloonzh/xueqiu-cli/xueqiu"
```

### 查询单个 symbol

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

### 查询多个 symbols

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

### Go 库中的 Chrome remote 模式

```go
provider := xueqiu.NewChromeCookieProvider(xueqiu.ChromeConfig{
	RemoteURL: "http://127.0.0.1:9222",
	Timeout:   30 * time.Second,
})
client := xueqiu.NewClient(http.DefaultClient, provider)
```

`Client.Quote` 返回 `xueqiu.Quote`：

```go
type Quote map[string]any
```

这个 map 保留雪球上游接口的完整响应：

```text
data.market
data.others
data.quote
data.tags
error_code
error_description
```

数值字段会解码成 `json.Number`，避免 `amount`、`volume`、时间戳等大数字被 `float64` 转换导致精度损失。

更多 Go 库示例见 [docs/library-usage.md](docs/library-usage.md)。

## 发布

推送 tag 后会自动构建并发布二进制：

```bash
git tag 1.0.0
git push origin 1.0.0
```

workflow 定义在 [.github/workflows/release.yml](.github/workflows/release.yml)。

