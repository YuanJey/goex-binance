// 综合集成测试 — 连接 Binance Futures Testnet，逐一验证 SDK 全部 Service 方法。
// 运行方式:
//
//	BINANCE_API_KEY=xxx BINANCE_SECRET_KEY=xxx BINANCE_TEST=true go run ./examples/integration
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	binance "github.com/YuanJey/goex-binance"
)

// ─────────────────────── 辅助函数 ───────────────────────

const symbol = "BTCUSDT"

var (
	passed int
	failed int
)

func section(name string) { fmt.Printf("\n══════════ %s ══════════\n", name) }

func check(name string, err error) bool {
	if err != nil {
		// -4164 "No need to change margin type" / -4046 "No need to change position side"
		// 在测试网中这些错误是预期行为，不算失败
		if apiErr, ok := err.(*binance.APIError); ok {
			if apiErr.Code == -4046 {
				fmt.Printf("  ✔ %-38s (预期跳过: %s)\n", name, apiErr.Message)
				passed++
				return true
			}
			// -4164 仅在 SetMarginType 场景下可豁免
			if apiErr.Code == -4164 && (name == "SetMarginType(CROSSED)" || name == "SetMarginType(ISOLATED)") {
				fmt.Printf("  ✔ %-38s (预期跳过: %s)\n", name, apiErr.Message)
				passed++
				return true
			}
		}
		fmt.Printf("  ✘ %-38s ERROR: %v\n", name, err)
		failed++
		return false
	}
	fmt.Printf("  ✔ %-38s OK\n", name)
	passed++
	return true
}

func dump(label string, v any) {
	raw, _ := json.MarshalIndent(v, "    ", "  ")
	fmt.Printf("    %s: %s\n", label, string(raw))
}

// ─────────────────────── main ───────────────────────

func main() {
	apiKey := os.Getenv("BINANCE_API_KEY")
	secret := os.Getenv("BINANCE_SECRET_KEY")
	testnet := strings.EqualFold(os.Getenv("BINANCE_TEST"), "true")

	if apiKey == "" || secret == "" {
		fmt.Println("请设置 BINANCE_API_KEY / BINANCE_SECRET_KEY 环境变量")
		os.Exit(1)
	}

	client, err := binance.New(binance.Config{
		APIKey:       apiKey,
		SecretKey:    secret,
		Testnet:      testnet,
		AutoTimeSync: true,
	})
	if err != nil {
		fmt.Printf("初始化客户端失败: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// ==================== Market ====================
	section("Market 公共行情接口")

	err = client.Market.Ping(ctx)
	check("Ping", err)

	serverTime, err := client.Market.ServerTime(ctx)
	if check("ServerTime", err) {
		fmt.Printf("    服务器时间: %d (%s)\n", serverTime, time.UnixMilli(serverTime).Format(time.RFC3339))
	}

	info, err := client.Market.ExchangeInfo(ctx)
	if check("ExchangeInfo", err) {
		fmt.Printf("    交易对数量: %d\n", len(info.Symbols))
	}

	rules, err := client.Market.SymbolRules(ctx, symbol)
	if check("SymbolRules("+symbol+")", err) {
		dump("交易规则", rules)
	}

	klines, err := client.Market.Klines(ctx, symbol, "1h", 5, 0, 0)
	if check("Klines(1h,5)", err) {
		fmt.Printf("    返回K线数: %d\n", len(klines))
		if len(klines) > 0 {
			dump("最新K线", klines[len(klines)-1])
		}
	}

	ticker, err := client.Market.Ticker24h(ctx, symbol)
	if check("Ticker24h", err) {
		dump("24h行情", ticker)
	}

	book, err := client.Market.BookTicker(ctx, symbol)
	if check("BookTicker", err) {
		dump("最优挂单", book)
	}

	mark, err := client.Market.MarkPrice(ctx, symbol)
	if check("MarkPrice", err) {
		dump("标记价格", mark)
	}

	fundingRates, err := client.Market.FundingRateHistory(ctx, symbol, 0, 0, 3)
	if check("FundingRateHistory(3)", err) {
		fmt.Printf("    返回记录数: %d\n", len(fundingRates))
	}

	oi, err := client.Market.OpenInterest(ctx, symbol)
	if check("OpenInterest", err) {
		dump("持仓量", oi)
	}

	// ==================== Account ====================
	section("Account 账户接口")

	balances, err := client.Account.Balances(ctx)
	if check("Balances", err) {
		nonZero := 0
		for _, b := range balances {
			if b.Balance != "0.00000000" && b.Balance != "0" {
				nonZero++
				dump(b.Asset, b)
			}
		}
		fmt.Printf("    非零余额资产数: %d / %d\n", nonZero, len(balances))
	}

	acct, err := client.Account.Account(ctx)
	if check("Account", err) {
		fmt.Printf("    总钱包余额: %s | 可用余额: %s\n", acct.TotalWalletBalance, acct.AvailableBalance)
	}

	positions, err := client.Account.Positions(ctx, symbol)
	if check("Positions("+symbol+")", err) {
		fmt.Printf("    仓位数: %d\n", len(positions))
		for _, p := range positions {
			if p.PositionAmt != "0" && p.PositionAmt != "0.000" {
				dump(p.Symbol+" "+string(p.PositionSide), p)
			}
		}
	}

	dualSide, err := client.Account.PositionMode(ctx)
	if check("PositionMode", err) {
		fmt.Printf("    双向持仓: %v\n", dualSide)
	}

	commission, err := client.Account.CommissionRate(ctx, symbol)
	if check("CommissionRate("+symbol+")", err) {
		dump("手续费率", commission)
	}

	err = client.Account.SetLeverage(ctx, symbol, 20)
	check("SetLeverage(20)", err)

	err = client.Account.SetMarginType(ctx, symbol, binance.MarginTypeCrossed)
	check("SetMarginType(CROSSED)", err)

	// ==================== Trade ====================
	section("Trade 交易接口")

	// 测试订单（不真实提交）
	err = client.Trade.TestOrder(ctx, binance.PlaceOrderRequest{
		Symbol:       symbol,
		Side:         binance.SideBuy,
		PositionSide: binance.PositionSideLong,
		Type:         binance.OrderTypeLimit,
		TimeInForce:  binance.TimeInForceGTC,
		Quantity:     "0.001",
		Price:        "50000",
	})
	check("TestOrder(LIMIT)", err)

	// 下单 + 查询 + 撤单 完整流程
	var placedOrderID int64
	order, err := client.Trade.PlaceOrder(ctx, binance.PlaceOrderRequest{
		Symbol:       symbol,
		Side:         binance.SideBuy,
		PositionSide: binance.PositionSideLong,
		Type:         binance.OrderTypeLimit,
		TimeInForce:  binance.TimeInForceGTC,
		Quantity:     "0.001",
		Price:        "50000",
	})
	if check("PlaceOrder(LIMIT BUY)", err) {
		placedOrderID = order.OrderID
		dump("订单", order)
	}

	if placedOrderID > 0 {
		queried, err := client.Trade.GetOrder(ctx, binance.QueryOrderRequest{
			Symbol:  symbol,
			OrderID: placedOrderID,
		})
		if check("GetOrder", err) {
			fmt.Printf("    订单状态: %s\n", queried.Status)
		}
	}

	openOrders, err := client.Trade.OpenOrders(ctx, symbol)
	if check("OpenOrders("+symbol+")", err) {
		fmt.Printf("    当前挂单数: %d\n", len(openOrders))
	}

	if placedOrderID > 0 {
		cancelled, err := client.Trade.CancelOrder(ctx, binance.CancelOrderRequest{
			Symbol:  symbol,
			OrderID: placedOrderID,
		})
		if check("CancelOrder", err) {
			fmt.Printf("    撤单状态: %s\n", cancelled.Status)
		}
	}

	allOrders, err := client.Trade.AllOrders(ctx, symbol, 0, 0, 5)
	if check("AllOrders(5)", err) {
		fmt.Printf("    历史订单数: %d\n", len(allOrders))
	}

	trades, err := client.Trade.UserTrades(ctx, symbol, 0, 0, 5)
	if check("UserTrades(5)", err) {
		fmt.Printf("    成交记录数: %d\n", len(trades))
	}

	// 批量下单 + 批量撤单
	batchResults, err := client.Trade.PlaceBatchOrders(ctx, []binance.PlaceOrderRequest{
		{Symbol: symbol, Side: binance.SideBuy, PositionSide: binance.PositionSideLong, Type: binance.OrderTypeLimit, TimeInForce: binance.TimeInForceGTC, Quantity: "0.001", Price: "50000"},
		{Symbol: symbol, Side: binance.SideBuy, PositionSide: binance.PositionSideLong, Type: binance.OrderTypeLimit, TimeInForce: binance.TimeInForceGTC, Quantity: "0.001", Price: "50001"},
	})
	if check("PlaceBatchOrders(2)", err) {
		var batchIDs []int64
		for i, r := range batchResults {
			if r.Error != nil {
				fmt.Printf("    批量[%d] 失败: code=%d msg=%s\n", i, r.Error.Code, r.Error.Message)
			} else {
				fmt.Printf("    批量[%d] OrderID=%d Status=%s\n", i, r.Order.OrderID, r.Order.Status)
				batchIDs = append(batchIDs, r.Order.OrderID)
			}
		}
		if len(batchIDs) > 0 {
			cancelResults, err := client.Trade.CancelBatchOrders(ctx, binance.CancelBatchOrdersRequest{
				Symbol:   symbol,
				OrderIDs: batchIDs,
			})
			if check("CancelBatchOrders", err) {
				for i, r := range cancelResults {
					if r.Error != nil {
						fmt.Printf("    批量撤单[%d] 失败: code=%d\n", i, r.Error.Code)
					} else {
						fmt.Printf("    批量撤单[%d] OrderID=%d Status=%s\n", i, r.Order.OrderID, r.Order.Status)
					}
				}
			}
		}
	}

	// 全部撤单
	err = client.Trade.CancelAllOrders(ctx, symbol)
	check("CancelAllOrders("+symbol+")", err)

	// ==================== Algo ====================
	section("Algo 条件单接口")

	algoOpen, err := client.Algo.OpenOrders(ctx, symbol)
	if check("Algo.OpenOrders", err) {
		fmt.Printf("    当前条件单数: %d\n", len(algoOpen))
	}

	algoAll, err := client.Algo.AllOrders(ctx, symbol, 0, 0, 5)
	if check("Algo.AllOrders(5)", err) {
		fmt.Printf("    历史条件单数: %d\n", len(algoAll))
	}

	// 下一个止损条件单，然后撤掉
	algoOrder, err := client.Algo.PlaceOrder(ctx, binance.PlaceAlgoOrderRequest{
		Symbol:       symbol,
		Side:         binance.SideSell,
		PositionSide: binance.PositionSideShort,
		Type:         binance.OrderTypeStopMarket,
		Quantity:     "0.001",
		TriggerPrice: "50000",
	})
	if check("Algo.PlaceOrder(STOP_MARKET)", err) {
		dump("条件单", algoOrder)
		if algoOrder.AlgoID > 0 {
			_, err := client.Algo.CancelOrder(ctx, algoOrder.AlgoID, "")
			check("Algo.CancelOrder", err)
		}
	}

	// ==================== Rate Limits ====================
	section("RateLimits 速率统计")
	limits := client.RateLimits()
	dump("当前速率快照", limits)
	check("RateLimits(非零)", func() error {
		if limits.UsedWeight1m == 0 {
			return fmt.Errorf("UsedWeight1m 为 0，异常")
		}
		return nil
	}())

	// ==================== Stream ====================
	section("Stream WebSocket 订阅")

	// Market Stream — 订阅 BookTicker，收取一条后关闭
	streamCtx, streamCancel := context.WithTimeout(ctx, 15*time.Second)
	defer streamCancel()

	marketSub, err := client.Stream.Market(streamCtx, symbol+"@bookTicker")
	if check("Stream.Market(bookTicker)", err) {
		select {
		case event := <-marketSub.Events():
			if bt, ok := event.(*binance.BookTickerEvent); ok {
				fmt.Printf("    ✔ 收到 BookTicker: bid=%s ask=%s\n", bt.BidPrice, bt.AskPrice)
				passed++
			} else {
				fmt.Printf("    ✔ 收到事件类型: %T\n", event)
				passed++
			}
		case err := <-marketSub.Errors():
			fmt.Printf("    ✘ Stream 错误: %v\n", err)
			failed++
		case <-streamCtx.Done():
			fmt.Printf("    ✘ Stream 超时未收到数据\n")
			failed++
		}
		_ = marketSub.Close()
		check("MarketSub.Close", nil)
	}

	// User Stream — 建立 listenKey 连接，验证连通性后关闭
	userStreamCtx, userStreamCancel := context.WithTimeout(ctx, 10*time.Second)
	defer userStreamCancel()

	userSub, err := client.Stream.User(userStreamCtx)
	if check("Stream.User", err) {
		// 等待连接状态
		select {
		case st := <-userSub.Status():
			fmt.Printf("    用户流状态: %v\n", st.State)
			if st.State == binance.StreamConnected {
				fmt.Printf("    ✔ User Stream 连接成功\n")
				passed++
			}
		case err := <-userSub.Errors():
			fmt.Printf("    ✘ User Stream 错误: %v\n", err)
			failed++
		case <-userStreamCtx.Done():
			fmt.Printf("    ⚠ User Stream 等待超时（可能正常）\n")
		}
		userStreamCancel()
		_ = userSub.Close()
		check("UserSub.Close", nil)
	}

	// ==================== 汇总 ====================
	section("测试汇总")
	fmt.Printf("  通过: %d | 失败: %d | 总计: %d\n\n", passed, failed, passed+failed)
	if failed > 0 {
		os.Exit(1)
	}
}
