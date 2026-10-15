package binance

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	MainnetRESTURL = "https://fapi.binance.com"
	MainnetWSURL   = "wss://fstream.binance.com/public/stream"
	MainnetUserWS  = "wss://fstream.binance.com/private/ws"
	TestnetRESTURL = "https://demo-fapi.binance.com"
	TestnetWSURL   = "wss://fstream.binancefuture.com/stream"
	TestnetUserWS  = "wss://fstream.binancefuture.com/ws"
)

// Logger is deliberately small so applications can adapt slog, zap, logrus, or
// any other logger without forcing that dependency on SDK users.
type Logger interface {
	Printf(format string, args ...any)
}

type Config struct {
	APIKey       string
	SecretKey    string
	Testnet      bool
	RESTURL      string
	MarketWSURL  string
	UserWSURL    string
	HTTPClient   *http.Client
	RecvWindow   time.Duration
	Logger       Logger
	UserAgent    string
	AutoTimeSync bool
}

func (c Config) withDefaults() (Config, error) {
	if c.RESTURL == "" {
		c.RESTURL = MainnetRESTURL
		if c.Testnet {
			c.RESTURL = TestnetRESTURL
		}
	}
	if c.MarketWSURL == "" {
		c.MarketWSURL = MainnetWSURL
		if c.Testnet {
			c.MarketWSURL = TestnetWSURL
		}
	}
	if c.UserWSURL == "" {
		c.UserWSURL = MainnetUserWS
		if c.Testnet {
			c.UserWSURL = TestnetUserWS
		}
	}
	c.RESTURL = strings.TrimRight(c.RESTURL, "/")
	c.MarketWSURL = strings.TrimRight(c.MarketWSURL, "/")
	c.UserWSURL = strings.TrimRight(c.UserWSURL, "/")
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}
	if c.RecvWindow == 0 {
		c.RecvWindow = 5 * time.Second
	}
	if c.RecvWindow < 0 || c.RecvWindow > 60*time.Second {
		return c, fmt.Errorf("binance: recvWindow must be between 0 and 60s")
	}
	if c.UserAgent == "" {
		c.UserAgent = "goex-binance/0.1"
	}
	return c, nil
}
