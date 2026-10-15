// Package rest implements Binance HTTP authentication and request transport.
// It is internal so transport details do not become part of the SDK API.
package rest

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/YuanJey/goex-binance/internal/model"
)

type Security uint8

const (
	Public Security = iota
	APIKey
	Signed
)

type Config struct {
	APIKey     string
	SecretKey  string
	RESTURL    string
	HTTPClient *http.Client
	RecvWindow time.Duration
	UserAgent  string
}

type Client struct {
	cfg          Config
	http         *http.Client
	clockOffset  atomic.Int64
	rateSnapshot atomic.Value
}

func New(cfg Config) *Client { return &Client{cfg: cfg, http: cfg.HTTPClient} }

func (r *Client) timestamp() int64 { return time.Now().UnixMilli() + r.clockOffset.Load() }

func sign(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func (r *Client) Do(ctx context.Context, method, path string, params url.Values, security Security, out any) error {
	if security != Public && r.cfg.APIKey == "" {
		return model.ErrCredentialsRequired
	}
	if security == Signed && r.cfg.SecretKey == "" {
		return model.ErrCredentialsRequired
	}
	base := cloneValues(params)
	err := r.doOnce(ctx, method, path, r.prepare(base, security), security, out)
	if security == Signed && model.IsAPIError(err, -1021) {
		if syncErr := r.SyncTime(ctx); syncErr != nil {
			return errors.Join(err, syncErr)
		}
		return r.doOnce(ctx, method, path, r.prepare(base, security), security, out)
	}
	return err
}

func (r *Client) prepare(params url.Values, security Security) url.Values {
	params = cloneValues(params)
	if security == Signed {
		params.Set("timestamp", strconv.FormatInt(r.timestamp(), 10))
		params.Set("recvWindow", strconv.FormatInt(r.cfg.RecvWindow.Milliseconds(), 10))
		params.Set("signature", sign(r.cfg.SecretKey, params.Encode()))
	}
	return params
}

func (r *Client) SyncTime(ctx context.Context) error {
	before := time.Now()
	var result struct {
		ServerTime int64 `json:"serverTime"`
	}
	if err := r.doOnce(ctx, http.MethodGet, "/fapi/v1/time", nil, Public, &result); err != nil {
		return err
	}
	after := time.Now()
	midpoint := before.Add(after.Sub(before) / 2).UnixMilli()
	r.clockOffset.Store(result.ServerTime - midpoint)
	return nil
}

func (r *Client) doOnce(ctx context.Context, method, path string, params url.Values, security Security, out any) error {
	encoded := params.Encode()
	endpoint := r.cfg.RESTURL + path
	var body io.Reader
	if method == http.MethodGet || method == http.MethodDelete {
		if encoded != "" {
			endpoint += "?" + encoded
		}
	} else {
		body = bytes.NewBufferString(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return fmt.Errorf("binance: create request: %w", err)
	}
	if method != http.MethodGet && method != http.MethodDelete {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if security != Public {
		req.Header.Set("X-MBX-APIKEY", r.cfg.APIKey)
	}
	req.Header.Set("User-Agent", r.cfg.UserAgent)

	resp, err := r.http.Do(req)
	if err != nil {
		wrapped := fmt.Errorf("binance: %s %s: %w", method, path, err)
		if method != http.MethodGet {
			return errors.Join(model.ErrOutcomeUnknown, wrapped)
		}
		return wrapped
	}
	defer resp.Body.Close()
	r.observeRateLimits(resp.Header)
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return fmt.Errorf("binance: read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &model.APIError{HTTPStatus: resp.StatusCode}
		_ = json.Unmarshal(raw, apiErr)
		if seconds, parseErr := strconv.Atoi(resp.Header.Get("Retry-After")); parseErr == nil {
			apiErr.RetryAfter = time.Duration(seconds) * time.Second
		}
		if apiErr.Message == "" {
			apiErr.Message = strings.TrimSpace(string(raw))
		}
		return apiErr
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("binance: decode %s %s response: %w", method, path, err)
	}
	return nil
}

func (r *Client) observeRateLimits(h http.Header) {
	s := model.RateLimitSnapshot{ObservedAt: time.Now()}
	s.UsedWeight1m, _ = strconv.Atoi(h.Get("X-MBX-USED-WEIGHT-1M"))
	s.OrderCount10s, _ = strconv.Atoi(h.Get("X-MBX-ORDER-COUNT-10S"))
	s.OrderCount1m, _ = strconv.Atoi(h.Get("X-MBX-ORDER-COUNT-1M"))
	r.rateSnapshot.Store(s)
}

func cloneValues(in url.Values) url.Values {
	if in == nil {
		return make(url.Values)
	}
	out := make(url.Values, len(in))
	for k, values := range in {
		out[k] = append([]string(nil), values...)
	}
	return out
}

func SetOptional(v url.Values, key, value string) {
	if value != "" {
		v.Set(key, value)
	}
}

func SetBool(v url.Values, key string, value *bool) {
	if value != nil {
		v.Set(key, strconv.FormatBool(*value))
	}
}

func (r *Client) RateLimits() model.RateLimitSnapshot {
	v := r.rateSnapshot.Load()
	if v == nil {
		return model.RateLimitSnapshot{}
	}
	return v.(model.RateLimitSnapshot)
}
