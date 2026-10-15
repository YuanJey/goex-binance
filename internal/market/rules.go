package market

import (
	"fmt"
	"math/big"

	"github.com/YuanJey/goex-binance/internal/model"
)

type Rules struct {
	Symbol                                     string
	MinPrice, MaxPrice, TickSize               string
	MinQty, MaxQty, StepSize                   string
	MarketMinQty, MarketMaxQty, MarketStepSize string
	MinNotional                                string
}

func newRules(info SymbolInfo) (*Rules, error) {
	r := &Rules{Symbol: info.Symbol}
	for _, f := range info.Filters {
		switch f.FilterType {
		case "PRICE_FILTER":
			r.MinPrice, r.MaxPrice, r.TickSize = f.MinPrice, f.MaxPrice, f.TickSize
		case "LOT_SIZE":
			r.MinQty, r.MaxQty, r.StepSize = f.MinQty, f.MaxQty, f.StepSize
		case "MARKET_LOT_SIZE":
			r.MarketMinQty, r.MarketMaxQty, r.MarketStepSize = f.MinQty, f.MaxQty, f.StepSize
		case "MIN_NOTIONAL":
			r.MinNotional = f.Notional
		}
	}
	if r.TickSize == "" || r.StepSize == "" {
		return nil, fmt.Errorf("binance: incomplete filters for %s", info.Symbol)
	}
	return r, nil
}

func (r *Rules) ValidatePrice(price string) error {
	return validateIncrement("price", price, r.MinPrice, r.MaxPrice, r.TickSize)
}

func (r *Rules) ValidateQuantity(qty string, market bool) error {
	min, max, step := r.MinQty, r.MaxQty, r.StepSize
	if market && r.MarketStepSize != "" {
		min, max, step = r.MarketMinQty, r.MarketMaxQty, r.MarketStepSize
	}
	return validateIncrement("quantity", qty, min, max, step)
}

func (r *Rules) NormalizePriceDown(price string) (string, error) {
	return floorToStep(price, r.TickSize)
}
func (r *Rules) NormalizeQuantityDown(qty string, market bool) (string, error) {
	step := r.StepSize
	if market && r.MarketStepSize != "" {
		step = r.MarketStepSize
	}
	return floorToStep(qty, step)
}

func validateIncrement(field, value, min, max, step string) error {
	v, ok := new(big.Rat).SetString(value)
	if !ok || v.Sign() < 0 {
		return &model.ValidationError{Field: field, Message: "invalid decimal"}
	}
	if min != "" && min != "0" {
		m, _ := new(big.Rat).SetString(min)
		if v.Cmp(m) < 0 {
			return &model.ValidationError{Field: field, Message: "below minimum " + min}
		}
	}
	if max != "" && max != "0" {
		m, _ := new(big.Rat).SetString(max)
		if v.Cmp(m) > 0 {
			return &model.ValidationError{Field: field, Message: "above maximum " + max}
		}
	}
	s, ok := new(big.Rat).SetString(step)
	if !ok || s.Sign() <= 0 {
		return &model.ValidationError{Field: field, Message: "invalid exchange step"}
	}
	q := new(big.Rat).Quo(v, s)
	if !q.IsInt() {
		return &model.ValidationError{Field: field, Message: "not aligned to step " + step}
	}
	return nil
}

func floorToStep(value, step string) (string, error) {
	v, ok := new(big.Rat).SetString(value)
	if !ok || v.Sign() < 0 {
		return "", &model.ValidationError{Message: "invalid decimal " + value}
	}
	s, ok := new(big.Rat).SetString(step)
	if !ok || s.Sign() <= 0 {
		return "", &model.ValidationError{Message: "invalid step " + step}
	}
	q := new(big.Rat).Quo(v, s)
	i := new(big.Int).Quo(q.Num(), q.Denom())
	result := new(big.Rat).Mul(new(big.Rat).SetInt(i), s)
	decimals := decimalPlaces(step)
	return result.FloatString(decimals), nil
}

func decimalPlaces(s string) int {
	for len(s) > 0 && s[len(s)-1] == '0' {
		s = s[:len(s)-1]
	}
	for i := range s {
		if s[i] == '.' {
			return len(s) - i - 1
		}
	}
	return 0
}
