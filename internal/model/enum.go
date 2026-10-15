// Package model contains values shared by the REST and WebSocket services.
package model

type Side string

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

type PositionSide string

const (
	PositionSideBoth  PositionSide = "BOTH"
	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"
)

type OrderType string

const (
	OrderTypeLimit              OrderType = "LIMIT"
	OrderTypeMarket             OrderType = "MARKET"
	OrderTypeStop               OrderType = "STOP"
	OrderTypeStopMarket         OrderType = "STOP_MARKET"
	OrderTypeTakeProfit         OrderType = "TAKE_PROFIT"
	OrderTypeTakeProfitMarket   OrderType = "TAKE_PROFIT_MARKET"
	OrderTypeTrailingStopMarket OrderType = "TRAILING_STOP_MARKET"
)

type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GTC"
	TimeInForceIOC TimeInForce = "IOC"
	TimeInForceFOK TimeInForce = "FOK"
	TimeInForceGTX TimeInForce = "GTX"
	TimeInForceGTD TimeInForce = "GTD"
	TimeInForceRPI TimeInForce = "RPI"
)

type WorkingType string

const (
	WorkingTypeContractPrice WorkingType = "CONTRACT_PRICE"
	WorkingTypeMarkPrice     WorkingType = "MARK_PRICE"
)

type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "NEW"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCanceled        OrderStatus = "CANCELED"
	OrderStatusExpired         OrderStatus = "EXPIRED"
	OrderStatusExpiredInMatch  OrderStatus = "EXPIRED_IN_MATCH"
)

type MarginType string

const (
	MarginTypeIsolated MarginType = "ISOLATED"
	MarginTypeCrossed  MarginType = "CROSSED"
)

type PriceMatch string
type SelfTradePreventionMode string
type NewOrderResponseType string

const (
	ResponseTypeACK    NewOrderResponseType    = "ACK"
	ResponseTypeResult NewOrderResponseType    = "RESULT"
	STPExpireTaker     SelfTradePreventionMode = "EXPIRE_TAKER"
	STPExpireMaker     SelfTradePreventionMode = "EXPIRE_MAKER"
	STPExpireBoth      SelfTradePreventionMode = "EXPIRE_BOTH"
)
