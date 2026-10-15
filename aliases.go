package binance

import (
	"github.com/YuanJey/goex-binance/internal/account"
	"github.com/YuanJey/goex-binance/internal/algo"
	"github.com/YuanJey/goex-binance/internal/market"
	"github.com/YuanJey/goex-binance/internal/model"
	"github.com/YuanJey/goex-binance/internal/stream"
	"github.com/YuanJey/goex-binance/internal/trade"
)

// Shared request and response types remain available from the root package for
// backwards compatibility. Their implementations live in focused packages.
type (
	Side                    = model.Side
	PositionSide            = model.PositionSide
	OrderType               = model.OrderType
	TimeInForce             = model.TimeInForce
	WorkingType             = model.WorkingType
	OrderStatus             = model.OrderStatus
	MarginType              = model.MarginType
	PriceMatch              = model.PriceMatch
	SelfTradePreventionMode = model.SelfTradePreventionMode
	NewOrderResponseType    = model.NewOrderResponseType
	APIError                = model.APIError
	ValidationError         = model.ValidationError
	RateLimitSnapshot       = model.RateLimitSnapshot

	AccountService  = account.Service
	Balance         = account.Balance
	Account         = account.Account
	AccountAsset    = account.AccountAsset
	AccountPosition = account.AccountPosition
	Position        = account.Position
	CommissionRate  = account.CommissionRate

	AlgoService           = algo.Service
	PlaceAlgoOrderRequest = algo.PlaceAlgoOrderRequest
	AlgoOrder             = algo.AlgoOrder

	MarketService = market.Service
	ExchangeInfo  = market.ExchangeInfo
	RateLimit     = market.RateLimit
	SymbolInfo    = market.SymbolInfo
	SymbolFilter  = market.SymbolFilter
	Kline         = market.Kline
	Ticker24h     = market.Ticker24h
	BookTicker    = market.BookTicker
	MarkPrice     = market.MarkPrice
	FundingRate   = market.FundingRate
	OpenInterest  = market.OpenInterest
	Rules         = market.Rules

	TradeService             = trade.Service
	PlaceOrderRequest        = trade.PlaceOrderRequest
	ModifyOrderRequest       = trade.ModifyOrderRequest
	Order                    = trade.Order
	CancelOrderRequest       = trade.CancelOrderRequest
	QueryOrderRequest        = trade.QueryOrderRequest
	UserTrade                = trade.UserTrade
	BatchOrderResult         = trade.BatchOrderResult
	CancelBatchOrdersRequest = trade.CancelBatchOrdersRequest

	StreamService        = stream.Service
	MarketSubscription   = stream.MarketSubscription
	UserSubscription     = stream.UserSubscription
	StreamState          = stream.StreamState
	StreamStatus         = stream.StreamStatus
	MarketEvent          = stream.MarketEvent
	AggTradeEvent        = stream.AggTradeEvent
	BookTickerEvent      = stream.BookTickerEvent
	KlineEvent           = stream.KlineEvent
	StreamKline          = stream.StreamKline
	MarkPriceEvent       = stream.MarkPriceEvent
	DepthEvent           = stream.DepthEvent
	RawMarketEvent       = stream.RawMarketEvent
	TickerEvent          = stream.TickerEvent
	MiniTickerEvent      = stream.MiniTickerEvent
	TickerArrayEvent     = stream.TickerArrayEvent
	MiniTickerArrayEvent = stream.MiniTickerArrayEvent
	UserEvent            = stream.UserEvent
	OrderTradeUpdate     = stream.OrderTradeUpdate
	UserOrderUpdate      = stream.UserOrderUpdate
	AccountUpdate        = stream.AccountUpdate
	UserAccountUpdate    = stream.UserAccountUpdate
	UserBalance          = stream.UserBalance
	UserPosition         = stream.UserPosition
	AccountConfigUpdate  = stream.AccountConfigUpdate
	MarginCallEvent      = stream.MarginCallEvent
	StrategyUpdate       = stream.StrategyUpdate
	AlgoUpdate           = stream.AlgoUpdate
	ListenKeyExpired     = stream.ListenKeyExpired
	RawUserEvent         = stream.RawUserEvent
)

const (
	SideBuy                     = model.SideBuy
	SideSell                    = model.SideSell
	PositionSideBoth            = model.PositionSideBoth
	PositionSideLong            = model.PositionSideLong
	PositionSideShort           = model.PositionSideShort
	OrderTypeLimit              = model.OrderTypeLimit
	OrderTypeMarket             = model.OrderTypeMarket
	OrderTypeStop               = model.OrderTypeStop
	OrderTypeStopMarket         = model.OrderTypeStopMarket
	OrderTypeTakeProfit         = model.OrderTypeTakeProfit
	OrderTypeTakeProfitMarket   = model.OrderTypeTakeProfitMarket
	OrderTypeTrailingStopMarket = model.OrderTypeTrailingStopMarket
	TimeInForceGTC              = model.TimeInForceGTC
	TimeInForceIOC              = model.TimeInForceIOC
	TimeInForceFOK              = model.TimeInForceFOK
	TimeInForceGTX              = model.TimeInForceGTX
	TimeInForceGTD              = model.TimeInForceGTD
	TimeInForceRPI              = model.TimeInForceRPI
	WorkingTypeContractPrice    = model.WorkingTypeContractPrice
	WorkingTypeMarkPrice        = model.WorkingTypeMarkPrice
	OrderStatusNew              = model.OrderStatusNew
	OrderStatusPartiallyFilled  = model.OrderStatusPartiallyFilled
	OrderStatusFilled           = model.OrderStatusFilled
	OrderStatusCanceled         = model.OrderStatusCanceled
	OrderStatusExpired          = model.OrderStatusExpired
	OrderStatusExpiredInMatch   = model.OrderStatusExpiredInMatch
	MarginTypeIsolated          = model.MarginTypeIsolated
	MarginTypeCrossed           = model.MarginTypeCrossed
	ResponseTypeACK             = model.ResponseTypeACK
	ResponseTypeResult          = model.ResponseTypeResult
	STPExpireTaker              = model.STPExpireTaker
	STPExpireMaker              = model.STPExpireMaker
	STPExpireBoth               = model.STPExpireBoth
	StreamConnecting            = stream.StreamConnecting
	StreamConnected             = stream.StreamConnected
	StreamReconnecting          = stream.StreamReconnecting
	StreamClosed                = stream.StreamClosed
)

var (
	ErrCredentialsRequired = model.ErrCredentialsRequired
	ErrOutcomeUnknown      = model.ErrOutcomeUnknown
	ErrStreamClosed        = model.ErrStreamClosed
	ErrSlowConsumer        = model.ErrSlowConsumer
)

func IsAPIError(err error, code int) bool { return model.IsAPIError(err, code) }

// Bool returns a pointer to v for optional boolean request fields.
func Bool(v bool) *bool { return &v }
