package binance

import (
	"context"

	"github.com/YuanJey/goex-binance/internal/account"
	"github.com/YuanJey/goex-binance/internal/algo"
	"github.com/YuanJey/goex-binance/internal/market"
	"github.com/YuanJey/goex-binance/internal/model"
	"github.com/YuanJey/goex-binance/internal/rest"
	"github.com/YuanJey/goex-binance/internal/stream"
	"github.com/YuanJey/goex-binance/internal/trade"
)

type Client struct {
	rest *rest.Client

	Market  *market.Service
	Account *account.Service
	Trade   *trade.Service
	Algo    *algo.Service
	Stream  *stream.Service
}

func New(cfg Config) (*Client, error) {
	cfg, err := cfg.withDefaults()
	if err != nil {
		return nil, err
	}
	r := rest.New(rest.Config{
		APIKey: cfg.APIKey, SecretKey: cfg.SecretKey, RESTURL: cfg.RESTURL,
		HTTPClient: cfg.HTTPClient, RecvWindow: cfg.RecvWindow, UserAgent: cfg.UserAgent,
	})
	c := &Client{rest: r}
	c.Market = market.New(r)
	c.Account = account.New(r)
	c.Trade = trade.New(r)
	c.Algo = algo.New(r)
	c.Stream = stream.New(r, stream.Config{APIKey: cfg.APIKey, MarketURL: cfg.MarketWSURL, UserURL: cfg.UserWSURL})
	if cfg.AutoTimeSync {
		if err := c.SyncTime(context.Background()); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *Client) SyncTime(ctx context.Context) error {
	return c.rest.SyncTime(ctx)
}

func (c *Client) RateLimits() model.RateLimitSnapshot { return c.rest.RateLimits() }
