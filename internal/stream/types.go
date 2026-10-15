package stream

import (
	"encoding/json"

	"github.com/YuanJey/goex-binance/internal/model"
)

type StreamState string

const (
	StreamConnecting   StreamState = "connecting"
	StreamConnected    StreamState = "connected"
	StreamReconnecting StreamState = "reconnecting"
	StreamClosed       StreamState = "closed"
)

type StreamStatus struct {
	State   StreamState
	Attempt int
	Err     error
}

type MarketEvent interface{ marketEvent() }

type AggTradeEvent struct {
	EventType        string `json:"e"`
	EventTime        int64  `json:"E"`
	Symbol           string `json:"s"`
	AggregateTradeID int64  `json:"a"`
	Price            string `json:"p"`
	Quantity         string `json:"q"`
	FirstTradeID     int64  `json:"f"`
	LastTradeID      int64  `json:"l"`
	TradeTime        int64  `json:"T"`
	BuyerIsMaker     bool   `json:"m"`
}

func (AggTradeEvent) marketEvent() {}

type BookTickerEvent struct {
	EventType       string `json:"e"`
	UpdateID        int64  `json:"u"`
	EventTime       int64  `json:"E"`
	TransactionTime int64  `json:"T"`
	Symbol          string `json:"s"`
	BidPrice        string `json:"b"`
	BidQuantity     string `json:"B"`
	AskPrice        string `json:"a"`
	AskQuantity     string `json:"A"`
}

func (BookTickerEvent) marketEvent() {}

type KlineEvent struct {
	EventType string      `json:"e"`
	EventTime int64       `json:"E"`
	Symbol    string      `json:"s"`
	Kline     StreamKline `json:"k"`
}
type StreamKline struct {
	StartTime           int64  `json:"t"`
	CloseTime           int64  `json:"T"`
	Symbol              string `json:"s"`
	Interval            string `json:"i"`
	FirstTradeID        int64  `json:"f"`
	LastTradeID         int64  `json:"L"`
	Open                string `json:"o"`
	Close               string `json:"c"`
	High                string `json:"h"`
	Low                 string `json:"l"`
	Volume              string `json:"v"`
	Trades              int64  `json:"n"`
	Closed              bool   `json:"x"`
	QuoteVolume         string `json:"q"`
	TakerBuyBaseVolume  string `json:"V"`
	TakerBuyQuoteVolume string `json:"Q"`
}

func (KlineEvent) marketEvent() {}

type MarkPriceEvent struct {
	EventType            string `json:"e"`
	EventTime            int64  `json:"E"`
	Symbol               string `json:"s"`
	MarkPrice            string `json:"p"`
	IndexPrice           string `json:"i"`
	EstimatedSettlePrice string `json:"P"`
	FundingRate          string `json:"r"`
	NextFundingTime      int64  `json:"T"`
}

func (MarkPriceEvent) marketEvent() {}

type DepthEvent struct {
	EventType             string     `json:"e"`
	EventTime             int64      `json:"E"`
	TransactionTime       int64      `json:"T"`
	Symbol                string     `json:"s"`
	FirstUpdateID         int64      `json:"U"`
	FinalUpdateID         int64      `json:"u"`
	PreviousFinalUpdateID int64      `json:"pu"`
	Bids                  [][]string `json:"b"`
	Asks                  [][]string `json:"a"`
}

func (DepthEvent) marketEvent() {}

type RawMarketEvent struct {
	Stream    string
	EventType string
	Data      json.RawMessage
}

func (RawMarketEvent) marketEvent() {}

type TickerEvent struct {
	EventType          string `json:"e"`
	EventTime          int64  `json:"E"`
	Symbol             string `json:"s"`
	PriceChange        string `json:"p"`
	PriceChangePercent string `json:"P"`
	WeightedAverage    string `json:"w"`
	LastPrice          string `json:"c"`
	LastQuantity       string `json:"Q"`
	OpenPrice          string `json:"o"`
	HighPrice          string `json:"h"`
	LowPrice           string `json:"l"`
	BaseVolume         string `json:"v"`
	QuoteVolume        string `json:"q"`
	OpenTime           int64  `json:"O"`
	CloseTime          int64  `json:"C"`
}

func (TickerEvent) marketEvent() {}

type MiniTickerEvent struct {
	EventType   string `json:"e"`
	EventTime   int64  `json:"E"`
	Symbol      string `json:"s"`
	ClosePrice  string `json:"c"`
	OpenPrice   string `json:"o"`
	HighPrice   string `json:"h"`
	LowPrice    string `json:"l"`
	BaseVolume  string `json:"v"`
	QuoteVolume string `json:"q"`
}

func (MiniTickerEvent) marketEvent() {}

type TickerArrayEvent []TickerEvent

func (TickerArrayEvent) marketEvent() {}

type MiniTickerArrayEvent []MiniTickerEvent

func (MiniTickerArrayEvent) marketEvent() {}

type UserEvent interface{ userEvent() }

type OrderTradeUpdate struct {
	EventType       string          `json:"e"`
	EventTime       int64           `json:"E"`
	TransactionTime int64           `json:"T"`
	Order           UserOrderUpdate `json:"o"`
}
type UserOrderUpdate struct {
	Symbol                    string             `json:"s"`
	ClientOrderID             string             `json:"c"`
	Side                      model.Side         `json:"S"`
	Type                      model.OrderType    `json:"o"`
	TimeInForce               model.TimeInForce  `json:"f"`
	OriginalQuantity          string             `json:"q"`
	OriginalPrice             string             `json:"p"`
	AveragePrice              string             `json:"ap"`
	StopPrice                 string             `json:"sp"`
	ExecutionType             string             `json:"x"`
	Status                    model.OrderStatus  `json:"X"`
	OrderID                   int64              `json:"i"`
	LastFilledQuantity        string             `json:"l"`
	AccumulatedFilledQuantity string             `json:"z"`
	LastFilledPrice           string             `json:"L"`
	CommissionAsset           string             `json:"N"`
	Commission                string             `json:"n"`
	OrderTradeTime            int64              `json:"T"`
	TradeID                   int64              `json:"t"`
	RealizedProfit            string             `json:"rp"`
	Maker                     bool               `json:"m"`
	ReduceOnly                bool               `json:"R"`
	WorkingType               model.WorkingType  `json:"wt"`
	OriginalType              model.OrderType    `json:"ot"`
	PositionSide              model.PositionSide `json:"ps"`
	ClosePosition             bool               `json:"cp"`
	ActivationPrice           string             `json:"AP"`
	CallbackRate              string             `json:"cr"`
}

func (OrderTradeUpdate) userEvent() {}

type AccountUpdate struct {
	EventType       string            `json:"e"`
	EventTime       int64             `json:"E"`
	TransactionTime int64             `json:"T"`
	Account         UserAccountUpdate `json:"a"`
}
type UserAccountUpdate struct {
	Reason    string         `json:"m"`
	Balances  []UserBalance  `json:"B"`
	Positions []UserPosition `json:"P"`
}
type UserBalance struct {
	Asset              string `json:"a"`
	WalletBalance      string `json:"wb"`
	CrossWalletBalance string `json:"cw"`
	BalanceChange      string `json:"bc"`
}
type UserPosition struct {
	Symbol              string             `json:"s"`
	PositionAmount      string             `json:"pa"`
	EntryPrice          string             `json:"ep"`
	BreakEvenPrice      string             `json:"bep"`
	AccumulatedRealized string             `json:"cr"`
	UnrealizedPnL       string             `json:"up"`
	MarginType          string             `json:"mt"`
	IsolatedWallet      string             `json:"iw"`
	PositionSide        model.PositionSide `json:"ps"`
}

func (AccountUpdate) userEvent() {}

type AccountConfigUpdate struct {
	EventType            string          `json:"e"`
	EventTime            int64           `json:"E"`
	TransactionTime      int64           `json:"T"`
	AccountConfiguration json.RawMessage `json:"ac"`
	SymbolConfiguration  json.RawMessage `json:"ai"`
}

func (AccountConfigUpdate) userEvent() {}

type MarginCallEvent struct {
	EventType          string            `json:"e"`
	EventTime          int64             `json:"E"`
	CrossWalletBalance string            `json:"cw"`
	Positions          []json.RawMessage `json:"p"`
}

func (MarginCallEvent) userEvent() {}

type StrategyUpdate struct {
	EventType       string          `json:"e"`
	EventTime       int64           `json:"E"`
	TransactionTime int64           `json:"T"`
	Strategy        json.RawMessage `json:"su"`
}

func (StrategyUpdate) userEvent() {}

type AlgoUpdate struct {
	EventType       string          `json:"e"`
	EventTime       int64           `json:"E"`
	TransactionTime int64           `json:"T"`
	Order           json.RawMessage `json:"o"`
}

func (AlgoUpdate) userEvent() {}

type ListenKeyExpired struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	ListenKey string `json:"listenKey"`
}

func (ListenKeyExpired) userEvent() {}

type RawUserEvent struct {
	EventType string
	Data      json.RawMessage
}

func (RawUserEvent) userEvent() {}
