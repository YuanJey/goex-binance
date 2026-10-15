package binance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestMarketStream(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		var request struct {
			Method string   `json:"method"`
			Params []string `json:"params"`
		}
		if err := conn.ReadJSON(&request); err != nil {
			return
		}
		if request.Method != "SUBSCRIBE" || len(request.Params) != 1 || request.Params[0] != "btcusdt@bookticker" {
			t.Errorf("request=%+v", request)
		}
		_ = conn.WriteJSON(map[string]any{"stream": "btcusdt@bookTicker", "data": map[string]any{"e": "bookTicker", "s": "BTCUSDT", "b": "60000", "B": "1", "a": "60001", "A": "2"}})
		<-time.After(100 * time.Millisecond)
	}))
	defer server.Close()
	c, _ := New(Config{MarketWSURL: "ws" + strings.TrimPrefix(server.URL, "http")})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	sub, err := c.Stream.Market(ctx, "BTCUSDT@bookTicker")
	if err != nil {
		t.Fatal(err)
	}
	select {
	case event := <-sub.Events():
		book, ok := event.(*BookTickerEvent)
		if !ok || book.BidPrice != "60000" {
			t.Fatalf("event=%T %+v", event, event)
		}
	case err := <-sub.Errors():
		t.Fatal(err)
	case <-ctx.Done():
		t.Fatal("timed out")
	}
	if err := sub.Close(); err != nil {
		t.Fatal(err)
	}
}
