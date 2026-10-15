# goex-binance

Binance USDⓈ-M 永续合约 Go SDK。保留 Binance 原生交易语义，覆盖公共行情、账户与仓位、普通订单、Algo 条件单，以及公共/私有 WebSocket 数据流。

## 特性

- REST HMAC-SHA256 签名与服务器时间自动校准
- 生产环境和 Demo/Testnet 环境
- 普通订单、批量下单/改单/撤单
- 独立的 Algo Order API（止盈止损、条件单、追踪止损）
- One-way Mode 与 Hedge Mode 原生字段
- `exchangeInfo` 精确价格/数量规则，不使用 `float64` 处理交易数值
- 公共行情和 User Data Stream 自动重连
- User Stream listen key 自动续期
- Binance API 类型化错误和限流计数
- 慢消费者检测；私有事件不会静默丢失

本项目仅支持 USDⓈ-M Futures `/fapi/*`，不包含现货、COIN-M、期权和 Portfolio Margin。

## 项目结构

```text
.
├── client.go, config.go    # 稳定的 SDK 入口与客户端装配
├── aliases.go              # 对外 API 兼容层
├── internal/
│   ├── account/            # 账户、仓位与杠杆
│   ├── algo/               # 条件单与 Algo Order
│   ├── market/             # 公共行情与交易规则
│   ├── trade/              # 普通订单与成交
│   ├── stream/             # 公共/私有 WebSocket
│   ├── rest/               # HTTP、签名、时间同步与限流
│   └── model/              # 模块共享枚举与错误
└── examples/               # 可运行示例
```

业务实现均放在 `internal` 下，调用方继续只依赖根包 `binance`；原有的类型名和调用方式保持不变。

## 安装

```bash
go get github.com/YuanJey/goex-binance
```

## 创建客户端

```go
client, err := binance.New(binance.Config{
    APIKey:       os.Getenv("BINANCE_API_KEY"),
    SecretKey:    os.Getenv("BINANCE_SECRET_KEY"),
    Testnet:      true,
    AutoTimeSync: true,
})
if err != nil {
    log.Fatal(err)
}
```

`APIKey` 和 `SecretKey` 只有访问账户或交易接口时才需要。公共行情客户端可以不提供凭证。

## 下单

价格和数量均使用 decimal string：

```go
order, err := client.Trade.PlaceOrder(ctx, binance.PlaceOrderRequest{
    Symbol:        "BTCUSDT",
    Side:          binance.SideBuy,
    PositionSide:  binance.PositionSideBoth,
    Type:          binance.OrderTypeLimit,
    TimeInForce:   binance.TimeInForceGTC,
    Quantity:      "0.001",
    Price:         "60000",
    ClientOrderID: "strategy-a-001",
})
```

Hedge Mode 下使用 `PositionSideLong` 或 `PositionSideShort`。Binance 不允许 Hedge Mode 订单携带 `reduceOnly`，SDK 会在发请求前拒绝这种组合。

写请求遇到网络错误时可能返回 `ErrOutcomeUnknown`。此时不要直接重复下单，应使用 `ClientOrderID` 查询订单状态：

```go
if errors.Is(err, binance.ErrOutcomeUnknown) {
    order, queryErr := client.Trade.GetOrder(ctx, binance.QueryOrderRequest{
        Symbol:        "BTCUSDT",
        ClientOrderID: "strategy-a-001",
    })
    _ = order
    _ = queryErr
}
```

## Algo 条件单

2026 版 Binance USDⓈ-M 将 TP/SL 和追踪止损放在独立的 Algo API 中：

```go
algo, err := client.Algo.PlaceOrder(ctx, binance.PlaceAlgoOrderRequest{
    Symbol:        "BTCUSDT",
    Side:          binance.SideSell,
    PositionSide:  binance.PositionSideLong,
    Type:          binance.OrderTypeStopMarket,
    TriggerPrice:  "58000",
    WorkingType:   binance.WorkingTypeMarkPrice,
    ClosePosition: binance.Bool(true),
    PriceProtect:  binance.Bool(true),
    ClientAlgoID:  "strategy-a-sl-001",
})
```

## 交易规则

```go
rules, err := client.Market.SymbolRules(ctx, "BTCUSDT")
if err != nil {
    log.Fatal(err)
}

price, err := rules.NormalizePriceDown("60123.456")
qty, err := rules.NormalizeQuantityDown("0.001234", false)

if err := rules.ValidatePrice(price); err != nil {
    log.Fatal(err)
}
if err := rules.ValidateQuantity(qty, false); err != nil {
    log.Fatal(err)
}
```

规范化必须由调用方显式执行；SDK 不会静默修改下单价格或数量。

## 公共 WebSocket

```go
sub, err := client.Stream.Market(
    ctx,
    "btcusdt@bookTicker",
    "btcusdt@markPrice@1s",
    "btcusdt@kline_1m",
)
if err != nil {
    log.Fatal(err)
}
defer sub.Close()

for {
    select {
    case event := <-sub.Events():
        switch e := event.(type) {
        case *binance.BookTickerEvent:
            fmt.Println(e.Symbol, e.BidPrice, e.AskPrice)
        case *binance.MarkPriceEvent:
            fmt.Println(e.Symbol, e.MarkPrice)
        case *binance.KlineEvent:
            fmt.Println(e.Symbol, e.Kline.Close)
        }
    case err := <-sub.Errors():
        log.Println("stream:", err)
    case status := <-sub.Status():
        log.Println("connection:", status.State)
    case <-ctx.Done():
        return
    }
}
```

公共行情通道满时保留较新的事件，并通过 `Errors()` 报告 `ErrSlowConsumer`。

## User Data Stream

```go
sub, err := client.Stream.User(ctx)
if err != nil {
    log.Fatal(err)
}
defer sub.Close()

for event := range sub.Events() {
    switch e := event.(type) {
    case *binance.OrderTradeUpdate:
        fmt.Println(e.Order.ClientOrderID, e.Order.Status)
    case *binance.AccountUpdate:
        fmt.Println(e.Account.Reason, e.Account.Positions)
    case *binance.AlgoUpdate:
        fmt.Println("algo update", string(e.Order))
    }
}
```

私有事件队列满时订阅会被关闭，并返回 `ErrSlowConsumer`，因为订单或账户事件不能静默丢弃。调用方应重新查询账户、持仓和未完成订单后再建立订阅。

## 错误处理

```go
var apiErr *binance.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.HTTPStatus, apiErr.Code, apiErr.Message)
}

if binance.IsAPIError(err, -1121) {
    // invalid symbol
}
```

最近一次响应中返回的限流计数可通过 `client.RateLimits()` 获取。

## 自定义环境

```go
client, err := binance.New(binance.Config{
    APIKey:      apiKey,
    SecretKey:   secret,
    RESTURL:     "https://custom-rest.example.com",
    MarketWSURL: "wss://custom-ws.example.com/public/stream",
    UserWSURL:   "wss://custom-ws.example.com/private/ws",
})
```

## 测试

```bash
go test ./...
go test -race ./...
go vet ./...
```

默认测试完全使用本地 HTTP/WebSocket 服务，不需要 API Key，也不会产生真实订单。

## 安全说明

- 不要把 API Key 写入代码或提交到 Git。
- 建议使用只允许 Futures Trade 的 API Key，并配置 IP 白名单。
- Demo/Testnet 与生产环境使用不同密钥。
- SDK 不会自动重试结果未知的写操作。
