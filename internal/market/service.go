// Package market implements public futures market-data endpoints.
package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/YuanJey/goex-binance/internal/model"
	"github.com/YuanJey/goex-binance/internal/rest"
)

type Service struct{ rest *rest.Client }

func New(client *rest.Client) *Service { return &Service{rest: client} }

type ExchangeInfo struct {
	Timezone   string       `json:"timezone"`
	ServerTime int64        `json:"serverTime"`
	RateLimits []RateLimit  `json:"rateLimits"`
	Symbols    []SymbolInfo `json:"symbols"`
}

type RateLimit struct {
	Type        string `json:"rateLimitType"`
	Interval    string `json:"interval"`
	IntervalNum int    `json:"intervalNum"`
	Limit       int    `json:"limit"`
}

type SymbolInfo struct {
	Symbol            string              `json:"symbol"`
	Pair              string              `json:"pair"`
	ContractType      string              `json:"contractType"`
	DeliveryDate      int64               `json:"deliveryDate"`
	OnboardDate       int64               `json:"onboardDate"`
	Status            string              `json:"status"`
	BaseAsset         string              `json:"baseAsset"`
	QuoteAsset        string              `json:"quoteAsset"`
	MarginAsset       string              `json:"marginAsset"`
	PricePrecision    int                 `json:"pricePrecision"`
	QuantityPrecision int                 `json:"quantityPrecision"`
	TriggerProtect    string              `json:"triggerProtect"`
	OrderTypes        []model.OrderType   `json:"orderTypes"`
	TimeInForce       []model.TimeInForce `json:"timeInForce"`
	Filters           []SymbolFilter      `json:"filters"`
}

type SymbolFilter struct {
	FilterType     string `json:"filterType"`
	MinPrice       string `json:"minPrice"`
	MaxPrice       string `json:"maxPrice"`
	TickSize       string `json:"tickSize"`
	MinQty         string `json:"minQty"`
	MaxQty         string `json:"maxQty"`
	StepSize       string `json:"stepSize"`
	Notional       string `json:"notional"`
	MultiplierUp   string `json:"multiplierUp"`
	MultiplierDown string `json:"multiplierDown"`
	Limit          int    `json:"limit"`
}

type Kline struct {
	OpenTime, CloseTime                     int64
	Open, High, Low, Close                  string
	Volume, QuoteVolume                     string
	Trades                                  int64
	TakerBuyBaseVolume, TakerBuyQuoteVolume string
}

type Ticker24h struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	LastPrice          string `json:"lastPrice"`
	OpenPrice          string `json:"openPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           int64  `json:"openTime"`
	CloseTime          int64  `json:"closeTime"`
}

type BookTicker struct {
	Symbol   string `json:"symbol"`
	BidPrice string `json:"bidPrice"`
	BidQty   string `json:"bidQty"`
	AskPrice string `json:"askPrice"`
	AskQty   string `json:"askQty"`
	Time     int64  `json:"time"`
}

type MarkPrice struct {
	Symbol               string `json:"symbol"`
	MarkPrice            string `json:"markPrice"`
	IndexPrice           string `json:"indexPrice"`
	EstimatedSettlePrice string `json:"estimatedSettlePrice"`
	LastFundingRate      string `json:"lastFundingRate"`
	InterestRate         string `json:"interestRate"`
	NextFundingTime      int64  `json:"nextFundingTime"`
	Time                 int64  `json:"time"`
}

type FundingRate struct {
	Symbol      string `json:"symbol"`
	FundingRate string `json:"fundingRate"`
	FundingTime int64  `json:"fundingTime"`
	MarkPrice   string `json:"markPrice"`
}

type OpenInterest struct {
	Symbol       string `json:"symbol"`
	OpenInterest string `json:"openInterest"`
	Time         int64  `json:"time"`
}

func (s *Service) Ping(ctx context.Context) error {
	return s.rest.Do(ctx, http.MethodGet, "/fapi/v1/ping", nil, rest.Public, &struct{}{})
}

func (s *Service) ServerTime(ctx context.Context) (int64, error) {
	var out struct {
		ServerTime int64 `json:"serverTime"`
	}
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v1/time", nil, rest.Public, &out)
	return out.ServerTime, err
}

func (s *Service) ExchangeInfo(ctx context.Context) (*ExchangeInfo, error) {
	var out ExchangeInfo
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v1/exchangeInfo", nil, rest.Public, &out)
	return &out, err
}

func (s *Service) SymbolRules(ctx context.Context, symbol string) (*Rules, error) {
	info, err := s.ExchangeInfo(ctx)
	if err != nil {
		return nil, err
	}
	for _, item := range info.Symbols {
		if item.Symbol == symbol {
			return newRules(item)
		}
	}
	return nil, &model.ValidationError{Field: "symbol", Message: "not found in exchangeInfo"}
}

func (s *Service) Klines(ctx context.Context, symbol, interval string, limit int, startTime, endTime int64) ([]Kline, error) {
	if symbol == "" || interval == "" {
		return nil, &model.ValidationError{Message: "symbol and interval are required"}
	}
	v := url.Values{"symbol": {symbol}, "interval": {interval}}
	if limit > 0 {
		v.Set("limit", strconv.Itoa(limit))
	}
	if startTime > 0 {
		v.Set("startTime", strconv.FormatInt(startTime, 10))
	}
	if endTime > 0 {
		v.Set("endTime", strconv.FormatInt(endTime, 10))
	}
	var raw [][]json.RawMessage
	if err := s.rest.Do(ctx, http.MethodGet, "/fapi/v1/klines", v, rest.Public, &raw); err != nil {
		return nil, err
	}
	out := make([]Kline, 0, len(raw))
	for i, row := range raw {
		if len(row) < 11 {
			return nil, fmt.Errorf("binance: malformed kline row %d", i)
		}
		var k Kline
		if err := decodeKline(row, &k); err != nil {
			return nil, fmt.Errorf("binance: decode kline row %d: %w", i, err)
		}
		out = append(out, k)
	}
	return out, nil
}

func decodeKline(row []json.RawMessage, k *Kline) error {
	fields := []any{&k.OpenTime, &k.Open, &k.High, &k.Low, &k.Close, &k.Volume, &k.CloseTime, &k.QuoteVolume, &k.Trades, &k.TakerBuyBaseVolume, &k.TakerBuyQuoteVolume}
	for i, target := range fields {
		if err := json.Unmarshal(row[i], target); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Ticker24h(ctx context.Context, symbol string) (*Ticker24h, error) {
	v := url.Values{"symbol": {symbol}}
	var out Ticker24h
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v1/ticker/24hr", v, rest.Public, &out)
	return &out, err
}

func (s *Service) BookTicker(ctx context.Context, symbol string) (*BookTicker, error) {
	v := url.Values{"symbol": {symbol}}
	var out BookTicker
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v1/ticker/bookTicker", v, rest.Public, &out)
	return &out, err
}

func (s *Service) MarkPrice(ctx context.Context, symbol string) (*MarkPrice, error) {
	v := url.Values{"symbol": {symbol}}
	var out MarkPrice
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v1/premiumIndex", v, rest.Public, &out)
	return &out, err
}

func (s *Service) FundingRateHistory(ctx context.Context, symbol string, startTime, endTime int64, limit int) ([]FundingRate, error) {
	v := url.Values{}
	rest.SetOptional(v, "symbol", symbol)
	if startTime > 0 {
		v.Set("startTime", strconv.FormatInt(startTime, 10))
	}
	if endTime > 0 {
		v.Set("endTime", strconv.FormatInt(endTime, 10))
	}
	if limit > 0 {
		v.Set("limit", strconv.Itoa(limit))
	}
	var out []FundingRate
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v1/fundingRate", v, rest.Public, &out)
	return out, err
}

func (s *Service) OpenInterest(ctx context.Context, symbol string) (*OpenInterest, error) {
	v := url.Values{"symbol": {symbol}}
	var out OpenInterest
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v1/openInterest", v, rest.Public, &out)
	return &out, err
}
