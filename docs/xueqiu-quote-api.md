# 雪球实时行情接口说明

本文档基于 `xueqiu-cli stock SH513310` 的实际 JSON 输出，并对照雪球页面展示内容整理。

## 接口来源

雪球页面：

```text
中韩半导体ETF华泰柏瑞(SH:513310)
¥3.923 +0.008 +0.20%
休市 04-24 15:00:00 北京时间
最高：3.969  今开：3.895  涨停：4.307  成交量：1614.41万手
最低：3.894  昨收：3.915  跌停：3.524  成交额：63.55亿
换手：--     市价：3.923  单位净值：3.688  基金份额：21.71亿
振幅：1.92%  溢价率：6.54%  累计净值：3.688  资产净值：84.79亿
成立日：2022-11-02  净值日期：2026-04-24  到期日：--  货币单位：CNY
```

上游接口：

```http
GET https://stock.xueqiu.com/v5/stock/quote.json?symbol=SH513310&extend=detail
```

请求需要带雪球 cookie。当前 CLI 会先通过 Chrome 打开 `https://xueqiu.com` 获取 cookie，再请求接口。

## CLI 用法

```bash
./xueqiu-cli stock SH513310
./xueqiu-cli stocks SH513310,SH000001
./xueqiu-cli --chrome-remote-url=http://127.0.0.1:9223 stock SH513310
```

`stock` 返回一个完整接口响应对象，`stocks` 返回完整接口响应对象数组。

## 响应结构

```json
{
  "data": {
    "market": {},
    "others": {},
    "quote": {},
    "tags": []
  },
  "error_code": 0,
  "error_description": ""
}
```

字段说明：

| 字段路径 | 类型 | 说明 |
| --- | --- | --- |
| `data.market` | object | 市场状态、时区、延迟信息 |
| `data.others` | object | 盘口等补充信息 |
| `data.quote` | object | 行情主体，包含价格、涨跌幅、成交、基金净值等字段 |
| `data.tags` | array | 页面标题附近的标签，例如融资、卖空、T+0 |
| `error_code` | number | 上游接口错误码，`0` 表示成功 |
| `error_description` | string | 上游接口错误描述 |

## 页面字段映射

以下映射以 `SH513310` 的实际 JSON 为例。

| 页面展示 | JSON 字段路径 | JSON 值示例 | 展示换算 |
| --- | --- | --- | --- |
| 名称 | `data.quote.name` | `中韩半导体ETF华泰柏瑞` | 直接展示 |
| Symbol | `data.quote.symbol` | `SH513310` | 页面展示为 `SH:513310` |
| 可融资 | `data.tags[].description` | `融` | 标签值直接展示 |
| 可卖空 | `data.tags[].description` | `空` | 标签值直接展示 |
| T+0交易 | `data.tags[].description` | `T+0交易` | 标签值直接展示 |
| 当前价 / 市价 | `data.quote.current` | `3.923` | 直接展示，页面加 `¥` |
| 涨跌额 | `data.quote.chg` | `0.008` | 正数展示 `+0.008` |
| 涨跌幅 | `data.quote.percent` | `0.2` | 百分数值，展示为 `+0.20%` |
| 市场状态 | `data.market.status` | `休市` | 直接展示 |
| 行情时间 | `data.quote.timestamp` 或 `data.quote.time` | `1777014000000` | 毫秒时间戳，北京时间 `2026-04-24 15:00:00` |
| 最高 | `data.quote.high` | `3.969` | 直接展示 |
| 今开 | `data.quote.open` | `3.895` | 直接展示 |
| 涨停 | `data.quote.limit_up` | `4.307` | 直接展示 |
| 成交量 | `data.quote.volume` | `1614405695` | 股数，页面按 `volume / 100 / 10000` 展示为 `1614.41万手` |
| 最低 | `data.quote.low` | `3.894` | 直接展示 |
| 昨收 | `data.quote.last_close` | `3.915` | 直接展示 |
| 跌停 | `data.quote.limit_down` | `3.524` | 直接展示 |
| 成交额 | `data.quote.amount` | `6355487484` | 元，页面按 `amount / 100000000` 展示为 `63.55亿` |
| 换手 | `data.quote.turnover_rate` | `null` | `null` 展示为 `--` |
| 单位净值 | `data.quote.unit_nav` | `3.688` | 直接展示 |
| 基金份额 | `data.quote.total_shares` | `2171427000` | 份，页面按 `total_shares / 100000000` 展示为 `21.71亿` |
| 振幅 | `data.quote.amplitude` | `1.92` | 百分数值，展示为 `1.92%` |
| 溢价率 | `data.quote.premium_rate` | `6.54` | 百分数值，展示为 `6.54%` |
| 累计净值 | `data.quote.acc_unit_nav` | `3.688` | 直接展示 |
| 资产净值 | `data.quote.market_capital` | `8479278121` | 元，页面按 `market_capital / 100000000` 展示为 `84.79亿` |
| 成立日 | `data.quote.found_date` | `1667318400000` | 毫秒时间戳，北京时间 `2022-11-02` |
| 净值日期 | `data.quote.nav_date` | `1776960000000` | 毫秒时间戳，北京时间 `2026-04-24` |
| 到期日 | `data.quote.expiration_date` | `null` | `null` 展示为 `--` |
| 货币单位 | `data.quote.currency` | `CNY` | 直接展示 |

页面上的 `2.55 万球友关注` 不在当前 `quote.json?extend=detail` 响应中。

## 常用字段说明

### `data.market`

| 字段 | 示例 | 说明 |
| --- | --- | --- |
| `status` | `休市` | 市场状态 |
| `status_id` | `8` | 市场状态 ID |
| `region` | `CN` | 市场区域 |
| `time_zone` | `Asia/Shanghai` | 市场时区 |
| `delay_tag` | `0` | 延迟标识，`0` 表示无延迟 |
| `daylight_savings` | `true` | 是否夏令时标识 |
| `downgrade_night_session` | `false` | 夜盘降级标识 |

### `data.quote`

`data.quote` 是主要行情对象。当前 CLI 会原样返回上游接口字段，不会裁剪字段。ETF 场景下常见字段包括：

| 字段 | 示例 | 说明 |
| --- | --- | --- |
| `symbol` | `SH513310` | 雪球 symbol |
| `code` | `513310` | 证券代码 |
| `name` | `中韩半导体ETF华泰柏瑞` | 名称 |
| `exchange` | `SH` | 交易所 |
| `currency` | `CNY` | 货币单位 |
| `current` | `3.923` | 当前价 |
| `chg` | `0.008` | 涨跌额 |
| `percent` | `0.2` | 涨跌幅，单位为百分数 |
| `open` | `3.895` | 今开 |
| `high` | `3.969` | 最高 |
| `low` | `3.894` | 最低 |
| `last_close` | `3.915` | 昨收 |
| `limit_up` | `4.307` | 涨停价 |
| `limit_down` | `3.524` | 跌停价 |
| `amount` | `6355487484` | 成交额，单位元 |
| `volume` | `1614405695` | 成交量，单位股 |
| `turnover_rate` | `null` | 换手率，百分数；为空时页面展示 `--` |
| `amplitude` | `1.92` | 振幅，单位为百分数 |
| `premium_rate` | `6.54` | 溢价率，单位为百分数 |
| `unit_nav` | `3.688` | 单位净值 |
| `acc_unit_nav` | `3.688` | 累计净值 |
| `iopv` | `3.6821` | ETF 实时参考净值 |
| `total_shares` | `2171427000` | 基金份额，单位份 |
| `market_capital` | `8479278121` | 资产净值，单位元 |
| `found_date` | `1667318400000` | 成立日，毫秒时间戳 |
| `nav_date` | `1776960000000` | 净值日期，毫秒时间戳 |
| `expiration_date` | `null` | 到期日；为空时页面展示 `--` |
| `timestamp` | `1777014000000` | 行情时间，毫秒时间戳 |
| `time` | `1777014000000` | 行情时间，通常与 `timestamp` 一致 |
| `type` | `13` | 证券类型 ID |
| `sub_type` | `EBS` | 证券子类型 |
| `lot_size` | `100` | 每手股数 |
| `tick_size` | `0.001` | 最小价格变动单位 |
| `delayed` | `0` | 延迟标识 |
| `status` | `1` | 标的状态 |

### `data.tags`

示例：

```json
[
  { "description": "融", "value": 6 },
  { "description": "空", "value": 7 },
  { "description": "T+0交易", "value": 8 }
]
```

这些字段对应页面标题后的能力标签：

| `description` | 页面含义 |
| --- | --- |
| `融` | 可融资 |
| `空` | 可卖空 |
| `T+0交易` | 支持 T+0 交易 |

## 展示规则

| 数据类型 | 规则 |
| --- | --- |
| 价格 | 直接展示数值，保留页面需要的小数位 |
| 百分比 | JSON 中已是百分数值，例如 `6.54` 展示为 `6.54%`，不要再乘以 `100` |
| 金额 | `amount`、`market_capital` 单位为元；展示“亿”时除以 `100000000` |
| 成交量 | `volume` 单位为股；展示“万手”时先除以 `100` 转为手，再除以 `10000` |
| 基金份额 | `total_shares` 单位为份；展示“亿”时除以 `100000000` |
| 时间 | 时间字段是毫秒时间戳；页面按北京时间 `Asia/Shanghai` 展示 |
| 空值 | `null` 对应页面上的 `--` |

