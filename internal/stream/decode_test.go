package stream

import (
	"testing"

	"github.com/YuanJey/goex-binance/internal/model"
)

func TestDecodeUserEvents(t *testing.T) {
	event, err := decodeUserEvent([]byte(`{"e":"ORDER_TRADE_UPDATE","E":1,"T":2,"o":{"s":"BTCUSDT","c":"cid","X":"FILLED","i":42}}`))
	if err != nil {
		t.Fatal(err)
	}
	update, ok := event.(*OrderTradeUpdate)
	if !ok || update.Order.OrderID != 42 || update.Order.Status != model.OrderStatusFilled {
		t.Fatalf("event=%T %+v", event, event)
	}
	algo, err := decodeUserEvent([]byte(`{"e":"ALGO_UPDATE","E":1,"o":{"s":"BTCUSDT"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := algo.(*AlgoUpdate); !ok {
		t.Fatalf("event=%T", algo)
	}
}
