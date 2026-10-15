package binance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestPlaceOrderSignsRequestAndCapturesLimits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fapi/v1/order" || r.Method != http.MethodPost {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("X-MBX-APIKEY"); got != "key" {
			t.Errorf("api key=%q", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		sig := r.Form.Get("signature")
		unsigned := cloneTestValues(r.Form)
		unsigned.Del("signature")
		if want := testSign("secret", unsigned.Encode()); sig != want {
			t.Errorf("signature=%s want %s", sig, want)
		}
		if r.Form.Get("symbol") != "BTCUSDT" || r.Form.Get("price") != "60000" {
			t.Errorf("unexpected form: %v", r.Form)
		}
		w.Header().Set("X-MBX-USED-WEIGHT-1M", "17")
		w.Header().Set("X-MBX-ORDER-COUNT-10S", "2")
		_, _ = w.Write([]byte(`{"symbol":"BTCUSDT","orderId":42,"status":"NEW","clientOrderId":"cid"}`))
	}))
	defer server.Close()
	c, err := New(Config{APIKey: "key", SecretKey: "secret", RESTURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	order, err := c.Trade.PlaceOrder(context.Background(), PlaceOrderRequest{Symbol: "BTCUSDT", Side: SideBuy, Type: OrderTypeLimit, TimeInForce: TimeInForceGTC, Quantity: "0.001", Price: "60000"})
	if err != nil {
		t.Fatal(err)
	}
	if order.OrderID != 42 {
		t.Fatalf("order id=%d", order.OrderID)
	}
	limits := c.RateLimits()
	if limits.UsedWeight1m != 17 || limits.OrderCount10s != 2 {
		t.Fatalf("limits=%+v", limits)
	}
}

func TestTimestampErrorSyncsAndRetries(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/fapi/v1/time":
			_ = json.NewEncoder(w).Encode(map[string]int64{"serverTime": time.Now().Add(2 * time.Second).UnixMilli()})
		case "/fapi/v3/balance":
			if attempts.Add(1) == 1 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"code":-1021,"msg":"outside recvWindow"}`))
				return
			}
			_, _ = w.Write([]byte(`[]`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()
	c, _ := New(Config{APIKey: "key", SecretKey: "secret", RESTURL: server.URL})
	if _, err := c.Account.Balances(context.Background()); err != nil {
		t.Fatal(err)
	}
	if attempts.Load() != 2 {
		t.Fatalf("attempts=%d", attempts.Load())
	}
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"code":-1121,"msg":"Invalid symbol."}`))
	}))
	defer server.Close()
	c, _ := New(Config{RESTURL: server.URL})
	_, err := c.Market.BookTicker(context.Background(), "NOPE")
	if !IsAPIError(err, -1121) {
		t.Fatalf("error=%T %v", err, err)
	}
}

func TestCustomRESTStillDefaultsWebSocketURLs(t *testing.T) {
	cfg, err := (Config{RESTURL: "http://localhost/"}).withDefaults()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RESTURL != "http://localhost" || cfg.MarketWSURL == "" || cfg.UserWSURL == "" {
		t.Fatalf("config=%+v", cfg)
	}
}

func valuesWithout(v url.Values, key string) string {
	c := cloneTestValues(v)
	c.Del(key)
	return strings.TrimSpace(c.Encode())
}

func testSign(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func cloneTestValues(in url.Values) url.Values {
	out := make(url.Values, len(in))
	for key, values := range in {
		out[key] = append([]string(nil), values...)
	}
	return out
}
