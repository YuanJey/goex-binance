// Package trade implements regular order endpoints.
package trade

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/YuanJey/goex-binance/internal/model"
	"github.com/YuanJey/goex-binance/internal/rest"
)

type Service struct{ rest *rest.Client }

func New(client *rest.Client) *Service { return &Service{rest: client} }

type PlaceOrderRequest struct {
	Symbol                  string
	Side                    model.Side
	PositionSide            model.PositionSide
	Type                    model.OrderType
	TimeInForce             model.TimeInForce
	Quantity                string
	Price                   string
	ClientOrderID           string
	ReduceOnly              *bool
	ResponseType            model.NewOrderResponseType
	PriceMatch              model.PriceMatch
	SelfTradePreventionMode model.SelfTradePreventionMode
	GoodTillDate            int64
}

type ModifyOrderRequest struct {
	Symbol        string
	Side          model.Side
	Quantity      string
	Price         string
	OrderID       int64
	ClientOrderID string
	PriceMatch    model.PriceMatch
}

type Order struct {
	Symbol           string             `json:"symbol"`
	Pair             string             `json:"pair"`
	OrderID          int64              `json:"orderId"`
	ClientOrderID    string             `json:"clientOrderId"`
	Status           model.OrderStatus  `json:"status"`
	Price            string             `json:"price"`
	AveragePrice     string             `json:"avgPrice"`
	OriginalQuantity string             `json:"origQty"`
	ExecutedQuantity string             `json:"executedQty"`
	CumulativeQuote  string             `json:"cumQuote"`
	TimeInForce      model.TimeInForce  `json:"timeInForce"`
	Type             model.OrderType    `json:"type"`
	OriginalType     model.OrderType    `json:"origType"`
	Side             model.Side         `json:"side"`
	PositionSide     model.PositionSide `json:"positionSide"`
	ReduceOnly       bool               `json:"reduceOnly"`
	ClosePosition    bool               `json:"closePosition"`
	StopPrice        string             `json:"stopPrice"`
	WorkingType      model.WorkingType  `json:"workingType"`
	PriceProtect     bool               `json:"priceProtect"`
	UpdateTime       int64              `json:"updateTime"`
}

type CancelOrderRequest struct {
	Symbol        string
	OrderID       int64
	ClientOrderID string
}

type QueryOrderRequest = CancelOrderRequest

type UserTrade struct {
	ID              int64              `json:"id"`
	OrderID         int64              `json:"orderId"`
	Symbol          string             `json:"symbol"`
	Side            model.Side         `json:"side"`
	PositionSide    model.PositionSide `json:"positionSide"`
	Price           string             `json:"price"`
	Quantity        string             `json:"qty"`
	QuoteQuantity   string             `json:"quoteQty"`
	Commission      string             `json:"commission"`
	CommissionAsset string             `json:"commissionAsset"`
	RealizedPnL     string             `json:"realizedPnl"`
	Maker           bool               `json:"maker"`
	Time            int64              `json:"time"`
}

// BatchOrderResult represents one item in a batch response. Binance may accept
// some items and reject others while returning HTTP 200 for the whole batch.
type BatchOrderResult struct {
	Order *Order
	Error *model.APIError
}

type CancelBatchOrdersRequest struct {
	Symbol         string
	OrderIDs       []int64
	ClientOrderIDs []string
}

func (r PlaceOrderRequest) values() (url.Values, error) {
	if r.Symbol == "" || r.Side == "" || r.Type == "" {
		return nil, &model.ValidationError{Message: "symbol, side and type are required"}
	}
	if r.Type == model.OrderTypeLimit && (r.Quantity == "" || (r.Price == "" && r.PriceMatch == "")) {
		return nil, &model.ValidationError{Message: "limit order requires quantity and price or priceMatch"}
	}
	if r.PositionSide != "" && r.PositionSide != model.PositionSideBoth && r.ReduceOnly != nil {
		return nil, &model.ValidationError{Field: "reduceOnly", Message: "cannot be sent in Hedge Mode"}
	}
	v := url.Values{"symbol": {r.Symbol}, "side": {string(r.Side)}, "type": {string(r.Type)}}
	rest.SetOptional(v, "positionSide", string(r.PositionSide))
	rest.SetOptional(v, "timeInForce", string(r.TimeInForce))
	rest.SetOptional(v, "quantity", r.Quantity)
	rest.SetOptional(v, "price", r.Price)
	rest.SetOptional(v, "newClientOrderId", r.ClientOrderID)
	rest.SetBool(v, "reduceOnly", r.ReduceOnly)
	rest.SetOptional(v, "newOrderRespType", string(r.ResponseType))
	rest.SetOptional(v, "priceMatch", string(r.PriceMatch))
	rest.SetOptional(v, "selfTradePreventionMode", string(r.SelfTradePreventionMode))
	if r.GoodTillDate > 0 {
		v.Set("goodTillDate", strconv.FormatInt(r.GoodTillDate, 10))
	}
	return v, nil
}

func (s *Service) PlaceOrder(ctx context.Context, req PlaceOrderRequest) (*Order, error) {
	v, err := req.values()
	if err != nil {
		return nil, err
	}
	var out Order
	err = s.rest.Do(ctx, http.MethodPost, "/fapi/v1/order", v, rest.Signed, &out)
	return &out, err
}

func (s *Service) TestOrder(ctx context.Context, req PlaceOrderRequest) error {
	v, err := req.values()
	if err != nil {
		return err
	}
	return s.rest.Do(ctx, http.MethodPost, "/fapi/v1/order/test", v, rest.Signed, nil)
}

func (s *Service) PlaceBatchOrders(ctx context.Context, requests []PlaceOrderRequest) ([]BatchOrderResult, error) {
	if len(requests) == 0 || len(requests) > 5 {
		return nil, &model.ValidationError{Field: "orders", Message: "batch must contain 1 to 5 orders"}
	}
	orders := make([]map[string]string, 0, len(requests))
	for _, req := range requests {
		v, err := req.values()
		if err != nil {
			return nil, err
		}
		item := make(map[string]string, len(v))
		for k := range v {
			item[k] = v.Get(k)
		}
		orders = append(orders, item)
	}
	raw, _ := json.Marshal(orders)
	v := url.Values{"batchOrders": {string(raw)}}
	var rawResponse []json.RawMessage
	if err := s.rest.Do(ctx, http.MethodPost, "/fapi/v1/batchOrders", v, rest.Signed, &rawResponse); err != nil {
		return nil, err
	}
	return decodeBatchResults(rawResponse)
}

func (s *Service) ModifyBatchOrders(ctx context.Context, requests []ModifyOrderRequest) ([]BatchOrderResult, error) {
	if len(requests) == 0 || len(requests) > 5 {
		return nil, &model.ValidationError{Field: "orders", Message: "batch must contain 1 to 5 orders"}
	}
	orders := make([]map[string]string, 0, len(requests))
	for _, req := range requests {
		if req.Symbol == "" || req.Side == "" || req.Quantity == "" || (req.Price == "" && req.PriceMatch == "") {
			return nil, &model.ValidationError{Message: "each modification requires symbol, side, quantity and price or priceMatch"}
		}
		if req.OrderID == 0 && req.ClientOrderID == "" {
			return nil, &model.ValidationError{Message: "each modification requires orderId or clientOrderId"}
		}
		item := map[string]string{"symbol": req.Symbol, "side": string(req.Side), "quantity": req.Quantity}
		if req.Price != "" {
			item["price"] = req.Price
		}
		if req.PriceMatch != "" {
			item["priceMatch"] = string(req.PriceMatch)
		}
		if req.OrderID != 0 {
			item["orderId"] = strconv.FormatInt(req.OrderID, 10)
		} else {
			item["origClientOrderId"] = req.ClientOrderID
		}
		orders = append(orders, item)
	}
	raw, _ := json.Marshal(orders)
	var rawResponse []json.RawMessage
	if err := s.rest.Do(ctx, http.MethodPut, "/fapi/v1/batchOrders", url.Values{"batchOrders": {string(raw)}}, rest.Signed, &rawResponse); err != nil {
		return nil, err
	}
	return decodeBatchResults(rawResponse)
}

func (s *Service) CancelBatchOrders(ctx context.Context, req CancelBatchOrdersRequest) ([]BatchOrderResult, error) {
	count := len(req.OrderIDs) + len(req.ClientOrderIDs)
	if req.Symbol == "" {
		return nil, &model.ValidationError{Field: "symbol", Message: "required"}
	}
	if count == 0 || count > 10 {
		return nil, &model.ValidationError{Message: "provide 1 to 10 order IDs"}
	}
	v := url.Values{"symbol": {req.Symbol}}
	if len(req.OrderIDs) > 0 {
		raw, _ := json.Marshal(req.OrderIDs)
		v.Set("orderIdList", string(raw))
	}
	if len(req.ClientOrderIDs) > 0 {
		raw, _ := json.Marshal(req.ClientOrderIDs)
		v.Set("origClientOrderIdList", string(raw))
	}
	var rawResponse []json.RawMessage
	if err := s.rest.Do(ctx, http.MethodDelete, "/fapi/v1/batchOrders", v, rest.Signed, &rawResponse); err != nil {
		return nil, err
	}
	return decodeBatchResults(rawResponse)
}

func decodeBatchResults(items []json.RawMessage) ([]BatchOrderResult, error) {
	out := make([]BatchOrderResult, 0, len(items))
	for _, raw := range items {
		var failure struct {
			Code    int    `json:"code"`
			Message string `json:"msg"`
		}
		if err := json.Unmarshal(raw, &failure); err != nil {
			return nil, err
		}
		if failure.Code != 0 {
			out = append(out, BatchOrderResult{Error: &model.APIError{HTTPStatus: http.StatusOK, Code: failure.Code, Message: failure.Message}})
			continue
		}
		var order Order
		if err := json.Unmarshal(raw, &order); err != nil {
			return nil, err
		}
		out = append(out, BatchOrderResult{Order: &order})
	}
	return out, nil
}

func (s *Service) ModifyOrder(ctx context.Context, req ModifyOrderRequest) (*Order, error) {
	if req.Symbol == "" || req.Side == "" || req.Quantity == "" || (req.Price == "" && req.PriceMatch == "") {
		return nil, &model.ValidationError{Message: "symbol, side, quantity and price or priceMatch are required"}
	}
	if req.OrderID == 0 && req.ClientOrderID == "" {
		return nil, &model.ValidationError{Message: "orderId or clientOrderId is required"}
	}
	v := url.Values{"symbol": {req.Symbol}, "side": {string(req.Side)}, "quantity": {req.Quantity}}
	rest.SetOptional(v, "price", req.Price)
	rest.SetOptional(v, "origClientOrderId", req.ClientOrderID)
	rest.SetOptional(v, "priceMatch", string(req.PriceMatch))
	if req.OrderID != 0 {
		v.Set("orderId", strconv.FormatInt(req.OrderID, 10))
	}
	var out Order
	err := s.rest.Do(ctx, http.MethodPut, "/fapi/v1/order", v, rest.Signed, &out)
	return &out, err
}

func (s *Service) CancelOrder(ctx context.Context, req CancelOrderRequest) (*Order, error) {
	v, err := orderIdentity(req.Symbol, req.OrderID, req.ClientOrderID)
	if err != nil {
		return nil, err
	}
	var out Order
	err = s.rest.Do(ctx, http.MethodDelete, "/fapi/v1/order", v, rest.Signed, &out)
	return &out, err
}

func (s *Service) CancelAllOrders(ctx context.Context, symbol string) error {
	if symbol == "" {
		return &model.ValidationError{Field: "symbol", Message: "required"}
	}
	return s.rest.Do(ctx, http.MethodDelete, "/fapi/v1/allOpenOrders", url.Values{"symbol": {symbol}}, rest.Signed, nil)
}

func (s *Service) GetOrder(ctx context.Context, req QueryOrderRequest) (*Order, error) {
	v, err := orderIdentity(req.Symbol, req.OrderID, req.ClientOrderID)
	if err != nil {
		return nil, err
	}
	var out Order
	err = s.rest.Do(ctx, http.MethodGet, "/fapi/v1/order", v, rest.Signed, &out)
	return &out, err
}

func (s *Service) OpenOrders(ctx context.Context, symbol string) ([]Order, error) {
	v := url.Values{}
	rest.SetOptional(v, "symbol", symbol)
	var out []Order
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v1/openOrders", v, rest.Signed, &out)
	return out, err
}

func (s *Service) AllOrders(ctx context.Context, symbol string, startTime, endTime int64, limit int) ([]Order, error) {
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
	var out []Order
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v1/allOrders", v, rest.Signed, &out)
	return out, err
}

func (s *Service) UserTrades(ctx context.Context, symbol string, startTime, endTime int64, limit int) ([]UserTrade, error) {
	if symbol == "" {
		return nil, &model.ValidationError{Field: "symbol", Message: "required"}
	}
	v := url.Values{"symbol": {symbol}}
	if startTime > 0 {
		v.Set("startTime", strconv.FormatInt(startTime, 10))
	}
	if endTime > 0 {
		v.Set("endTime", strconv.FormatInt(endTime, 10))
	}
	if limit > 0 {
		v.Set("limit", strconv.Itoa(limit))
	}
	var out []UserTrade
	err := s.rest.Do(ctx, http.MethodGet, "/fapi/v1/userTrades", v, rest.Signed, &out)
	return out, err
}

func orderIdentity(symbol string, orderID int64, clientOrderID string) (url.Values, error) {
	if symbol == "" {
		return nil, &model.ValidationError{Field: "symbol", Message: "required"}
	}
	if orderID == 0 && clientOrderID == "" {
		return nil, &model.ValidationError{Message: "orderId or clientOrderId is required"}
	}
	v := url.Values{"symbol": {symbol}}
	if orderID != 0 {
		v.Set("orderId", strconv.FormatInt(orderID, 10))
	} else {
		v.Set("origClientOrderId", clientOrderID)
	}
	return v, nil
}
