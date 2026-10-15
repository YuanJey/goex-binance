// Package stream implements public and private WebSocket subscriptions.
package stream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/YuanJey/goex-binance/internal/model"
	"github.com/YuanJey/goex-binance/internal/rest"
	"github.com/gorilla/websocket"
)

type Config struct {
	APIKey    string
	MarketURL string
	UserURL   string
}

type Service struct {
	rest *rest.Client
	cfg  Config
}

func New(client *rest.Client, cfg Config) *Service { return &Service{rest: client, cfg: cfg} }

type MarketSubscription struct {
	events chan MarketEvent
	errors chan error
	status chan StreamStatus
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

func (s *MarketSubscription) Events() <-chan MarketEvent  { return s.events }
func (s *MarketSubscription) Errors() <-chan error        { return s.errors }
func (s *MarketSubscription) Status() <-chan StreamStatus { return s.status }
func (s *MarketSubscription) Close() error                { s.once.Do(s.cancel); <-s.done; return nil }

type UserSubscription struct {
	events chan UserEvent
	errors chan error
	status chan StreamStatus
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

func (s *UserSubscription) Events() <-chan UserEvent    { return s.events }
func (s *UserSubscription) Errors() <-chan error        { return s.errors }
func (s *UserSubscription) Status() <-chan StreamStatus { return s.status }
func (s *UserSubscription) Close() error                { s.once.Do(s.cancel); <-s.done; return nil }

func (s *Service) Market(ctx context.Context, streams ...string) (*MarketSubscription, error) {
	if len(streams) == 0 {
		return nil, &model.ValidationError{Field: "streams", Message: "at least one stream is required"}
	}
	for i := range streams {
		streams[i] = strings.ToLower(strings.TrimSpace(streams[i]))
		if streams[i] == "" {
			return nil, &model.ValidationError{Field: "streams", Message: "stream name cannot be empty"}
		}
	}
	ctx, cancel := context.WithCancel(ctx)
	sub := &MarketSubscription{events: make(chan MarketEvent, 256), errors: make(chan error, 16), status: make(chan StreamStatus, 16), cancel: cancel, done: make(chan struct{})}
	go s.runMarket(ctx, sub, append([]string(nil), streams...))
	return sub, nil
}

func (s *Service) User(ctx context.Context) (*UserSubscription, error) {
	if s.cfg.APIKey == "" {
		return nil, model.ErrCredentialsRequired
	}
	ctx, cancel := context.WithCancel(ctx)
	sub := &UserSubscription{events: make(chan UserEvent, 1024), errors: make(chan error, 16), status: make(chan StreamStatus, 16), cancel: cancel, done: make(chan struct{})}
	go s.runUser(ctx, sub)
	return sub, nil
}

func (s *Service) runMarket(ctx context.Context, sub *MarketSubscription, streams []string) {
	defer close(sub.done)
	defer close(sub.events)
	defer close(sub.errors)
	defer close(sub.status)
	backoff := time.Second
	for attempt := 0; ; attempt++ {
		if ctx.Err() != nil {
			notifyStatus(sub.status, StreamStatus{State: StreamClosed, Err: ctx.Err()})
			return
		}
		state := StreamConnecting
		if attempt > 0 {
			state = StreamReconnecting
		}
		notifyStatus(sub.status, StreamStatus{State: state, Attempt: attempt})
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, s.cfg.MarketURL, nil)
		if err != nil {
			notifyErr(sub.errors, err)
			if !waitBackoff(ctx, backoff) {
				return
			}
			backoff = nextBackoff(backoff)
			continue
		}
		if err := conn.WriteJSON(map[string]any{"method": "SUBSCRIBE", "params": streams, "id": time.Now().UnixNano()}); err != nil {
			_ = conn.Close()
			notifyErr(sub.errors, err)
			continue
		}
		notifyStatus(sub.status, StreamStatus{State: StreamConnected, Attempt: attempt})
		backoff = time.Second
		err = readMarket(ctx, conn, sub)
		_ = conn.Close()
		if ctx.Err() != nil {
			notifyStatus(sub.status, StreamStatus{State: StreamClosed, Err: ctx.Err()})
			return
		}
		notifyErr(sub.errors, fmt.Errorf("binance market stream disconnected: %w", err))
		if !waitBackoff(ctx, backoff) {
			return
		}
		backoff = nextBackoff(backoff)
	}
}

func readMarket(ctx context.Context, conn *websocket.Conn, sub *MarketSubscription) error {
	stop := closeOnContext(ctx, conn)
	defer stop()
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Minute))
	defaultPing := conn.PingHandler()
	conn.SetPingHandler(func(appData string) error {
		_ = conn.SetReadDeadline(time.Now().Add(10 * time.Minute))
		return defaultPing(appData)
	})
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		var envelope struct {
			Stream string          `json:"stream"`
			Data   json.RawMessage `json:"data"`
		}
		data := json.RawMessage(raw)
		stream := ""
		if json.Unmarshal(raw, &envelope) == nil && len(envelope.Data) > 0 {
			data, stream = envelope.Data, envelope.Stream
		}
		event, err := decodeMarketEvent(stream, data)
		if err != nil {
			notifyErr(sub.errors, err)
			continue
		}
		publishLatest(sub.events, event, sub.errors)
	}
}

func decodeMarketEvent(stream string, data json.RawMessage) (MarketEvent, error) {
	if len(data) > 0 && data[0] == '[' {
		if strings.Contains(strings.ToLower(stream), "miniticker") {
			var events MiniTickerArrayEvent
			if err := json.Unmarshal(data, &events); err != nil {
				return nil, err
			}
			return events, nil
		}
		var events TickerArrayEvent
		if err := json.Unmarshal(data, &events); err != nil {
			return nil, err
		}
		return events, nil
	}
	var head struct {
		EventType string `json:"e"`
		EventTime int64  `json:"E"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return nil, fmt.Errorf("binance: decode market event: %w", err)
	}
	var target MarketEvent
	switch {
	case head.EventType == "aggTrade" || strings.Contains(stream, "@aggtrade"):
		target = &AggTradeEvent{}
	case head.EventType == "bookTicker" || strings.Contains(stream, "@bookticker"):
		target = &BookTickerEvent{}
	case head.EventType == "kline" || strings.Contains(stream, "@kline_"):
		target = &KlineEvent{}
	case head.EventType == "markPriceUpdate" || strings.Contains(stream, "@markprice"):
		target = &MarkPriceEvent{}
	case head.EventType == "24hrTicker" || strings.Contains(stream, "@ticker"):
		target = &TickerEvent{}
	case head.EventType == "24hrMiniTicker" || strings.Contains(stream, "@miniticker"):
		target = &MiniTickerEvent{}
	case head.EventType == "depthUpdate" || strings.Contains(stream, "@depth"):
		target = &DepthEvent{}
	default:
		return RawMarketEvent{Stream: stream, EventType: head.EventType, Data: append(json.RawMessage(nil), data...)}, nil
	}
	if err := json.Unmarshal(data, target); err != nil {
		return nil, err
	}
	return target, nil
}

var errListenKeyExpired = errors.New("binance: listen key expired")

func (s *Service) runUser(ctx context.Context, sub *UserSubscription) {
	defer close(sub.done)
	defer close(sub.events)
	defer close(sub.errors)
	defer close(sub.status)
	backoff := time.Second
	attempt := 0
	for ctx.Err() == nil {
		key, err := s.startListenKey(ctx)
		if err != nil {
			notifyErr(sub.errors, err)
			if !waitBackoff(ctx, backoff) {
				return
			}
			backoff = nextBackoff(backoff)
			continue
		}
		for ctx.Err() == nil {
			state := StreamConnecting
			if attempt > 0 {
				state = StreamReconnecting
			}
			notifyStatus(sub.status, StreamStatus{State: state, Attempt: attempt})
			conn, _, err := websocket.DefaultDialer.DialContext(ctx, s.cfg.UserURL+"/"+key, http.Header{"X-MBX-APIKEY": []string{s.cfg.APIKey}})
			if err != nil {
				notifyErr(sub.errors, err)
				attempt++
				if !waitBackoff(ctx, backoff) {
					break
				}
				backoff = nextBackoff(backoff)
				continue
			}
			notifyStatus(sub.status, StreamStatus{State: StreamConnected, Attempt: attempt})
			backoff = time.Second
			err = s.readUser(ctx, conn, sub)
			_ = conn.Close()
			if errors.Is(err, errListenKeyExpired) {
				break
			}
			if ctx.Err() == nil {
				notifyErr(sub.errors, fmt.Errorf("binance user stream disconnected: %w", err))
				attempt++
				if !waitBackoff(ctx, backoff) {
					break
				}
				backoff = nextBackoff(backoff)
			}
		}
		_ = s.closeListenKey(context.Background())
	}
	notifyStatus(sub.status, StreamStatus{State: StreamClosed, Err: ctx.Err()})
}

func (s *Service) readUser(ctx context.Context, conn *websocket.Conn, sub *UserSubscription) error {
	readCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := closeOnContext(readCtx, conn)
	defer stop()
	keepaliveErr := make(chan error, 1)
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-readCtx.Done():
				return
			case <-ticker.C:
				if err := s.keepaliveListenKey(readCtx); err != nil {
					keepaliveErr <- err
					_ = conn.Close()
					return
				}
			}
		}
	}()
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Minute))
	defaultPing := conn.PingHandler()
	conn.SetPingHandler(func(appData string) error {
		_ = conn.SetReadDeadline(time.Now().Add(10 * time.Minute))
		return defaultPing(appData)
	})
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			select {
			case keepErr := <-keepaliveErr:
				return keepErr
			default:
				return err
			}
		}
		event, err := decodeUserEvent(raw)
		if err != nil {
			notifyErr(sub.errors, err)
			continue
		}
		if _, expired := event.(*ListenKeyExpired); expired {
			return errListenKeyExpired
		}
		select {
		case sub.events <- event:
		default:
			notifyErr(sub.errors, model.ErrSlowConsumer)
			sub.cancel()
			return model.ErrSlowConsumer
		}
	}
}

func decodeUserEvent(data json.RawMessage) (UserEvent, error) {
	var head struct {
		EventType string `json:"e"`
		EventTime int64  `json:"E"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return nil, err
	}
	var target UserEvent
	switch head.EventType {
	case "ORDER_TRADE_UPDATE":
		target = &OrderTradeUpdate{}
	case "ACCOUNT_UPDATE":
		target = &AccountUpdate{}
	case "ACCOUNT_CONFIG_UPDATE":
		target = &AccountConfigUpdate{}
	case "MARGIN_CALL":
		target = &MarginCallEvent{}
	case "STRATEGY_UPDATE":
		target = &StrategyUpdate{}
	case "ALGO_UPDATE":
		target = &AlgoUpdate{}
	case "listenKeyExpired":
		target = &ListenKeyExpired{}
	default:
		return RawUserEvent{EventType: head.EventType, Data: append(json.RawMessage(nil), data...)}, nil
	}
	if err := json.Unmarshal(data, target); err != nil {
		return nil, err
	}
	return target, nil
}

func (s *Service) startListenKey(ctx context.Context) (string, error) {
	var out struct {
		ListenKey string `json:"listenKey"`
	}
	err := s.rest.Do(ctx, http.MethodPost, "/fapi/v1/listenKey", nil, rest.APIKey, &out)
	return out.ListenKey, err
}
func (s *Service) keepaliveListenKey(ctx context.Context) error {
	return s.rest.Do(ctx, http.MethodPut, "/fapi/v1/listenKey", nil, rest.APIKey, nil)
}
func (s *Service) closeListenKey(ctx context.Context) error {
	return s.rest.Do(ctx, http.MethodDelete, "/fapi/v1/listenKey", nil, rest.APIKey, nil)
}

func closeOnContext(ctx context.Context, conn *websocket.Conn) func() {
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()
	return func() { close(done) }
}
func waitBackoff(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
func nextBackoff(d time.Duration) time.Duration {
	d *= 2
	if d > 30*time.Second {
		return 30 * time.Second
	}
	return d
}
func notifyErr(ch chan error, err error) {
	if err == nil {
		return
	}
	select {
	case ch <- err:
	default:
	}
}
func notifyStatus(ch chan StreamStatus, status StreamStatus) {
	select {
	case ch <- status:
	default:
	}
}
func publishLatest(ch chan MarketEvent, event MarketEvent, errs chan error) {
	select {
	case ch <- event:
	default:
		select {
		case <-ch:
		default:
		}
		select {
		case ch <- event:
		default:
		}
		notifyErr(errs, model.ErrSlowConsumer)
	}
}
