package market

import "testing"

func TestRulesExactDecimalValidationAndNormalization(t *testing.T) {
	r := &Rules{MinPrice: "0.10", MaxPrice: "100000", TickSize: "0.10", MinQty: "0.001", MaxQty: "100", StepSize: "0.001"}
	if err := r.ValidatePrice("123.40"); err != nil {
		t.Fatal(err)
	}
	if err := r.ValidatePrice("123.45"); err == nil {
		t.Fatal("expected tick validation error")
	}
	if err := r.ValidateQuantity("0.002", false); err != nil {
		t.Fatal(err)
	}
	got, err := r.NormalizeQuantityDown("1.23456", false)
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.234" {
		t.Fatalf("normalized=%q", got)
	}
}

func TestNewRules(t *testing.T) {
	r, err := newRules(SymbolInfo{Symbol: "BTCUSDT", Filters: []SymbolFilter{{FilterType: "PRICE_FILTER", MinPrice: "0.1", MaxPrice: "100", TickSize: "0.1"}, {FilterType: "LOT_SIZE", MinQty: "0.001", MaxQty: "10", StepSize: "0.001"}, {FilterType: "MIN_NOTIONAL", Notional: "5"}}})
	if err != nil {
		t.Fatal(err)
	}
	if r.MinNotional != "5" {
		t.Fatalf("rules=%+v", r)
	}
}
