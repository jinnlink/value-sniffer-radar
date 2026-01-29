package marketdata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type EastmoneyRepoProvider struct {
	name        string
	baseURL     string
	fields      string
	rateDivisor float64
	httpClient  *http.Client
}

type EastmoneyRepoOptions struct {
	Name        string
	BaseURL     string
	Fields      string
	RateDivisor float64
	Timeout     time.Duration
}

func NewEastmoneyRepo(opt EastmoneyRepoOptions) *EastmoneyRepoProvider {
	name := strings.TrimSpace(opt.Name)
	if name == "" {
		name = "eastmoney_repo"
	}
	baseURL := strings.TrimSpace(opt.BaseURL)
	if baseURL == "" {
		baseURL = "https://push2.eastmoney.com/api/qt/stock/get"
	}
	fields := strings.TrimSpace(opt.Fields)
	if fields == "" {
		// f43 is "latest price" in Eastmoney quote responses; for repo we treat it as rate (%).
		fields = "f43,f57,f58,f59"
	}
	div := opt.RateDivisor
	if div == 0 {
		div = 1.0
	}
	to := opt.Timeout
	if to <= 0 {
		to = 1500 * time.Millisecond
	}
	return &EastmoneyRepoProvider{
		name:        name,
		baseURL:     baseURL,
		fields:      fields,
		rateDivisor: div,
		httpClient: &http.Client{
			Timeout: to,
		},
	}
}

func (p *EastmoneyRepoProvider) Name() string { return p.name }

func (p *EastmoneyRepoProvider) Fetch(ctx context.Context, symbol string) (Snapshot, error) {
	secid, err := eastmoneySecID(symbol)
	if err != nil {
		return Snapshot{Provider: p.name, Symbol: symbol, TS: time.Now()}, err
	}

	u, err := url.Parse(p.baseURL)
	if err != nil {
		return Snapshot{Provider: p.name, Symbol: symbol, TS: time.Now()}, err
	}
	q := u.Query()
	q.Set("secid", secid)
	q.Set("fields", p.fields)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Snapshot{Provider: p.name, Symbol: symbol, TS: time.Now()}, err
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return Snapshot{Provider: p.name, Symbol: symbol, TS: time.Now()}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Snapshot{Provider: p.name, Symbol: symbol, TS: time.Now()}, fmt.Errorf("eastmoney http status=%s", resp.Status)
	}

	var payload struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Snapshot{Provider: p.name, Symbol: symbol, TS: time.Now()}, err
	}
	if payload.Data == nil {
		return Snapshot{Provider: p.name, Symbol: symbol, TS: time.Now()}, errors.New("eastmoney missing data")
	}

	rawRate, ok := payload.Data["f43"]
	if !ok {
		return Snapshot{Provider: p.name, Symbol: symbol, TS: time.Now(), Raw: payload.Data}, errors.New("eastmoney missing f43")
	}
	rate, ok := anyToFloat(rawRate)
	if !ok {
		return Snapshot{Provider: p.name, Symbol: symbol, TS: time.Now(), Raw: payload.Data}, errors.New("eastmoney invalid f43")
	}
	if p.rateDivisor != 0 {
		rate = rate / p.rateDivisor
	}

	return Snapshot{
		Provider: p.name,
		Symbol:   symbol,
		TS:       time.Now(),
		RatePct:  rate,
		Raw:      payload.Data,
	}, nil
}

func eastmoneySecID(symbol string) (string, error) {
	s := strings.TrimSpace(strings.ToUpper(symbol))
	if s == "" {
		return "", errors.New("empty symbol")
	}
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid symbol: %s", symbol)
	}
	code, suffix := parts[0], parts[1]
	if len(code) != 6 {
		return "", fmt.Errorf("invalid code: %s", symbol)
	}
	switch suffix {
	case "SH":
		return "1." + code, nil
	case "SZ":
		return "0." + code, nil
	default:
		return "", fmt.Errorf("unknown market suffix: %s", symbol)
	}
}

func anyToFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	case string:
		// Some Eastmoney fields may render as strings.
		var n json.Number = json.Number(strings.TrimSpace(t))
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

