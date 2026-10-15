// Package algo implements conditional and trigger order endpoints.
package algo

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

type PlaceAlgoOrderRequest struct {
	Symbol                  string
	Side                    model.Side
	PositionSide            model.PositionSide
	Type                    model.OrderType
	TimeInForce             model.TimeInForce
	Quantity                string
	Price                   string
	TriggerPrice            string
	WorkingType             model.WorkingType
	PriceMatch              model.PriceMatch
	ClosePosition           *bool
	PriceProtect            *bool
	ReduceOnly              *bool
	ClientAlgoID            string
	CallbackRate            string
	ActivationPrice         string
	SelfTradePreventionMode model.SelfTradePreventionMode
	GoodTillDate            int64
}

type AlgoOrder struct {
	AlgoID         int64              `json:"algoId"`
	ClientAlgoID   string             `json:"clientAlgoId"`
	AlgoType       string             `json:"algoType"`
	OrderType      model.OrderType    `json:"orderType"`
	Symbol         string             `json:"symbol"`
	Side           model.Side         `json:"side"`
	PositionSide   model.PositionSide `json:"positionSide"`
	TimeInForce    model.TimeInForce  `json:"timeInForce"`
	Quantity       string             `json:"quantity"`
	AlgoStatus     string             `json:"algoStatus"`
	ActualOrderID  string             `json:"actualOrderId"`
	ActualPrice    string             `json:"actualPrice"`
	ActualType     model.OrderType    `json:"actualType"`
	ActualQuantity string             `json:"actualQty"`
	TriggerPrice   string             `json:"triggerPrice"`
	Price          string             `json:"price"`
	WorkingType    model.WorkingType  `json:"workingType"`
	ClosePosition  bool               `json:"closePosition"`
	PriceProtect   bool               `json:"priceProtect"`
	ReduceOnly     bool               `json:"reduceOnly"`
	CreateTime     int64              `json:"createTime"`
	UpdateTime     int64              `json:"updateTime"`
	TriggerTime    int64              `json:"triggerTime"`
}

func (r PlaceAlgoOrderRequest) values() (url.Values, error) {
	switch r.Type {
	case model.OrderTypeStop, model.OrderTypeStopMarket, model.OrderTypeTakeProfit, model.OrderTypeTakeProfitMarket, model.OrderTypeTrailingStopMarket:
	default:
		return nil, &model.ValidationError{Field: "type", Message: "unsupported algo order type"}
	}
	if r.Symbol == "" || r.Side == "" {
		return nil, &model.ValidationError{Message: "symbol and side are required"}
	}
	if r.Type != model.OrderTypeTrailingStopMarket && r.TriggerPrice == "" {
		return nil, &model.ValidationError{Field: "triggerPrice", Message: "required"}
	}
	v := url.Values{"algoType": {"CONDITIONAL"}, "symbol": {r.Symbol}, "side": {string(r.Side)}, "type": {string(r.Type)}}
	rest.SetOptional(v, "positionSide", string(r.PositionSide))
	rest.SetOptional(v, "timeInForce", string(r.TimeInForce))
	rest.SetOptional(v, "quantity", r.Quantity)
	rest.SetOptional(v, "price", r.Price)
	rest.SetOptional(v, "triggerPrice", r.TriggerPrice)
	rest.SetOptional(v, "workingType", string(r.WorkingType))
	rest.SetOptional(v, "priceMatch", string(r.PriceMatch))
	rest.SetBool(v, "closePosition", r.ClosePosition)
	rest.SetBool(v, "priceProtect", r.PriceProtect)
	rest.SetBool(v, "reduceOnly", r.ReduceOnly)
	rest.SetOptional(v, "clientAlgoId", r.ClientAlgoID)
	rest.SetOptional(v, "callbackRate", r.CallbackRate)
	rest.SetOptional(v, "activationPrice", r.ActivationPrice)
	rest.SetOptional(v, "selfTradePreventionMode", string(r.SelfTradePreventionMode))
	if r.GoodTillDate > 0 {
		v.Set("goodTillDate", strconv.FormatInt(r.GoodTillDate, 10))
	}
	return v, nil
}

func (s *Service) PlaceOrder(ctx context.Context, req PlaceAlgoOrderRequest) (*AlgoOrder, error) {
	v, err := req.values()
	if err != nil {
		return nil, err
	}
	var out AlgoOrder
	err = s.rest.Do(ctx, http.MethodPost, "/fapi/v1/algoOrder", v, rest.Signed, &out)
	return &out, err
}

func (s *Service) GetOrder(ctx context.Context, algoID int64, clientAlgoID string) (*AlgoOrder, error) {
	v, err := algoIdentity(algoID, clientAlgoID)
	if err != nil {
		return nil, err
	}
	var out AlgoOrder
	err = s.rest.Do(ctx, http.MethodGet, "/fapi/v1/algoOrder", v, rest.Signed, &out)
	return &out, err
}

func (s *Service) CancelOrder(ctx context.Context, algoID int64, clientAlgoID string) (*AlgoOrder, error) {
	v, err := algoIdentity(algoID, clientAlgoID)
	if err != nil {
		return nil, err
	}
	var out AlgoOrder
	err = s.rest.Do(ctx, http.MethodDelete, "/fapi/v1/algoOrder", v, rest.Signed, &out)
	return &out, err
}

func (s *Service) OpenOrders(ctx context.Context, symbol string) ([]AlgoOrder, error) {
	v := url.Values{}
	rest.SetOptional(v, "symbol", symbol)
	var out []AlgoOrder
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v1/openAlgoOrders", v, rest.Signed, &out)
	return out, err
}

func (s *Service) AllOrders(ctx context.Context, symbol string, startTime, endTime int64, limit int) ([]AlgoOrder, error) {
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
	var out []AlgoOrder
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v1/allAlgoOrders", v, rest.Signed, &out)
	return out, err
}

func (s *Service) CancelAllOpenOrders(ctx context.Context, symbol string) error {
	if symbol == "" {
		return &model.ValidationError{Field: "symbol", Message: "required"}
	}
	return s.rest.Do(ctx, http.MethodDelete, "/fapi/v1/algoOpenOrders", url.Values{"symbol": {symbol}}, rest.Signed, nil)
}

func algoIdentity(algoID int64, clientAlgoID string) (url.Values, error) {
	if algoID == 0 && clientAlgoID == "" {
		return nil, &model.ValidationError{Message: "algoId or clientAlgoId is required"}
	}
	v := url.Values{}
	if algoID != 0 {
		v.Set("algoId", strconv.FormatInt(algoID, 10))
	} else {
		v.Set("clientAlgoId", clientAlgoID)
	}
	return v, nil
}
