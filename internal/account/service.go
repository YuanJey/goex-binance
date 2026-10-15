// Package account implements account and position endpoints.
package account

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/YuanJey/goex-binance/internal/model"
	"github.com/YuanJey/goex-binance/internal/rest"
)

type Service struct{ rest *rest.Client }

func New(client *rest.Client) *Service { return &Service{rest: client} }

type Balance struct {
	AccountAlias       string `json:"accountAlias"`
	Asset              string `json:"asset"`
	Balance            string `json:"balance"`
	CrossWalletBalance string `json:"crossWalletBalance"`
	CrossUnPnl         string `json:"crossUnPnl"`
	AvailableBalance   string `json:"availableBalance"`
	MaxWithdrawAmount  string `json:"maxWithdrawAmount"`
	MarginAvailable    bool   `json:"marginAvailable"`
	UpdateTime         int64  `json:"updateTime"`
}

type Account struct {
	FeeTier                     int               `json:"feeTier"`
	CanTrade                    bool              `json:"canTrade"`
	CanDeposit                  bool              `json:"canDeposit"`
	CanWithdraw                 bool              `json:"canWithdraw"`
	UpdateTime                  int64             `json:"updateTime"`
	TotalInitialMargin          string            `json:"totalInitialMargin"`
	TotalMaintMargin            string            `json:"totalMaintMargin"`
	TotalWalletBalance          string            `json:"totalWalletBalance"`
	TotalUnrealizedProfit       string            `json:"totalUnrealizedProfit"`
	TotalMarginBalance          string            `json:"totalMarginBalance"`
	TotalPositionInitialMargin  string            `json:"totalPositionInitialMargin"`
	TotalOpenOrderInitialMargin string            `json:"totalOpenOrderInitialMargin"`
	AvailableBalance            string            `json:"availableBalance"`
	MaxWithdrawAmount           string            `json:"maxWithdrawAmount"`
	Assets                      []AccountAsset    `json:"assets"`
	Positions                   []AccountPosition `json:"positions"`
}

type AccountAsset struct {
	Asset                  string `json:"asset"`
	WalletBalance          string `json:"walletBalance"`
	UnrealizedProfit       string `json:"unrealizedProfit"`
	MarginBalance          string `json:"marginBalance"`
	MaintMargin            string `json:"maintMargin"`
	InitialMargin          string `json:"initialMargin"`
	PositionInitialMargin  string `json:"positionInitialMargin"`
	OpenOrderInitialMargin string `json:"openOrderInitialMargin"`
	MaxWithdrawAmount      string `json:"maxWithdrawAmount"`
	CrossWalletBalance     string `json:"crossWalletBalance"`
	CrossUnPnl             string `json:"crossUnPnl"`
	AvailableBalance       string `json:"availableBalance"`
	MarginAvailable        bool   `json:"marginAvailable"`
	UpdateTime             int64  `json:"updateTime"`
}

type AccountPosition struct {
	Symbol           string             `json:"symbol"`
	PositionSide     model.PositionSide `json:"positionSide"`
	PositionAmt      string             `json:"positionAmt"`
	UnrealizedProfit string             `json:"unrealizedProfit"`
	IsolatedMargin   string             `json:"isolatedMargin"`
	Notional         string             `json:"notional"`
	IsolatedWallet   string             `json:"isolatedWallet"`
	InitialMargin    string             `json:"initialMargin"`
	MaintMargin      string             `json:"maintMargin"`
	UpdateTime       int64              `json:"updateTime"`
}

type Position struct {
	Symbol           string             `json:"symbol"`
	PositionAmt      string             `json:"positionAmt"`
	EntryPrice       string             `json:"entryPrice"`
	BreakEvenPrice   string             `json:"breakEvenPrice"`
	MarkPrice        string             `json:"markPrice"`
	UnrealizedProfit string             `json:"unRealizedProfit"`
	LiquidationPrice string             `json:"liquidationPrice"`
	Leverage         string             `json:"leverage"`
	MaxNotionalValue string             `json:"maxNotionalValue"`
	MarginType       string             `json:"marginType"`
	IsolatedMargin   string             `json:"isolatedMargin"`
	IsAutoAddMargin  string             `json:"isAutoAddMargin"`
	PositionSide     model.PositionSide `json:"positionSide"`
	Notional         string             `json:"notional"`
	IsolatedWallet   string             `json:"isolatedWallet"`
	UpdateTime       int64              `json:"updateTime"`
}

type CommissionRate struct {
	Symbol              string `json:"symbol"`
	MakerCommissionRate string `json:"makerCommissionRate"`
	TakerCommissionRate string `json:"takerCommissionRate"`
}

func (s *Service) Balances(ctx context.Context) ([]Balance, error) {
	var out []Balance
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v3/balance", nil, rest.Signed, &out)
	return out, err
}

func (s *Service) Account(ctx context.Context) (*Account, error) {
	var out Account
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v3/account", nil, rest.Signed, &out)
	return &out, err
}

func (s *Service) Positions(ctx context.Context, symbol string) ([]Position, error) {
	v := url.Values{}
	rest.SetOptional(v, "symbol", symbol)
	var out []Position
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v3/positionRisk", v, rest.Signed, &out)
	return out, err
}

func (s *Service) PositionMode(ctx context.Context) (bool, error) {
	var out struct {
		DualSidePosition bool `json:"dualSidePosition"`
	}
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v1/positionSide/dual", nil, rest.Signed, &out)
	return out.DualSidePosition, err
}

func (s *Service) SetPositionMode(ctx context.Context, hedgeMode bool) error {
	v := url.Values{"dualSidePosition": {strconv.FormatBool(hedgeMode)}}
	return s.rest.Do(ctx, http.MethodPost, "/fapi/v1/positionSide/dual", v, rest.Signed, nil)
}

func (s *Service) SetLeverage(ctx context.Context, symbol string, leverage int) error {
	if leverage < 1 {
		return &model.ValidationError{Field: "leverage", Message: "must be positive; the maximum is symbol-specific"}
	}
	v := url.Values{"symbol": {symbol}, "leverage": {strconv.Itoa(leverage)}}
	return s.rest.Do(ctx, http.MethodPost, "/fapi/v1/leverage", v, rest.Signed, nil)
}

func (s *Service) SetMarginType(ctx context.Context, symbol string, marginType model.MarginType) error {
	v := url.Values{"symbol": {symbol}, "marginType": {string(marginType)}}
	return s.rest.Do(ctx, http.MethodPost, "/fapi/v1/marginType", v, rest.Signed, nil)
}

func (s *Service) ModifyPositionMargin(ctx context.Context, symbol string, side model.PositionSide, amount string, add bool) error {
	typeValue := "2"
	if add {
		typeValue = "1"
	}
	v := url.Values{"symbol": {symbol}, "amount": {amount}, "type": {typeValue}}
	if side != "" {
		v.Set("positionSide", string(side))
	}
	return s.rest.Do(ctx, http.MethodPost, "/fapi/v1/positionMargin", v, rest.Signed, nil)
}

func (s *Service) CommissionRate(ctx context.Context, symbol string) (*CommissionRate, error) {
	v := url.Values{"symbol": {symbol}}
	var out CommissionRate
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v1/commissionRate", v, rest.Signed, &out)
	return &out, err
}
