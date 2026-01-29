package marketdata

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TencentRepoProvider struct {
	name       string
	quoteURL   string
	httpClient *http.Client
}

type TencentRepoOptions struct {
	Name     string
	QuoteURL string
	Timeout  time.Duration
}

func NewTencentRepo(opt TencentRepoOptions) *TencentRepoProvider {
	name := strings.TrimSpace(opt.Name)
	if name == "" {
		name = "tencent_repo"
	}
	quoteURL := strings.TrimSpace(opt.QuoteURL)
	if quoteURL == "" {
		quoteURL = "https://qt.gtimg.cn/q="
	}
	to := opt.Timeout
	if to <= 0 {
		to = 1500 * time.Millisecond
	}
	return &TencentRepoProvider{
		name:     name,
		quoteURL: quoteURL,
		httpClient: &http.Client{
			Timeout: to,
		},
	}
}

func (p *TencentRepoProvider) Name() string { return p.name }

func (p *TencentRepoProvider) Fetch(ctx context.Context, symbol string) (Snapshot, error) {
	code, err := tencentCode(symbol)
	if err != nil {
		return Snapshot{Provider: p.name, Symbol: symbol, TS: time.Now()}, err
	}

	base := p.quoteURL
	if !strings.Contains(base, "q=") {
		// allow passing "https://qt.gtimg.cn/q=" or "https://qt.gtimg.cn/"
		if strings.HasSuffix(base, "/") {
			base += "q="
		} else if strings.HasSuffix(base, "q=") {
			// ok
		} else if strings.HasSuffix(base, "?") {
			base += "q="
		} else {
			if strings.Contains(base, "?") {
				base += "&q="
			} else {
				base += "?q="
			}
		}
	}

	u, err := url.Parse(base + url.QueryEscape(code))
	if err != nil {
		return Snapshot{Provider: p.name, Symbol: symbol, TS: time.Now()}, err
	}

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
		return Snapshot{Provider: p.name, Symbol: symbol, TS: time.Now()}, fmt.Errorf("tencent http status=%s", resp.Status)
	}

	// Response is plain text:
	// v_sh204001="...~...~...~<price>~...";
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		rate, raw, err := parseTencentLine(line)
		if err != nil {
			continue
		}
		return Snapshot{
			Provider: p.name,
			Symbol:   symbol,
			TS:       time.Now(),
			RatePct:  rate,
			Raw: map[string]any{
				"line":   line,
				"fields": raw,
			},
		}, nil
	}
	if err := sc.Err(); err != nil {
		return Snapshot{Provider: p.name, Symbol: symbol, TS: time.Now()}, err
	}
	return Snapshot{Provider: p.name, Symbol: symbol, TS: time.Now()}, errors.New("tencent empty response")
}

func tencentCode(symbol string) (string, error) {
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
		return "sh" + code, nil
	case "SZ":
		return "sz" + code, nil
	default:
		return "", fmt.Errorf("unknown market suffix: %s", symbol)
	}
}

func parseTencentLine(line string) (float64, []string, error) {
	i := strings.Index(line, "\"")
	j := strings.LastIndex(line, "\"")
	if i < 0 || j <= i {
		return 0, nil, errors.New("no quoted payload")
	}
	payload := line[i+1 : j]
	parts := strings.Split(payload, "~")
	// For many Tencent quotes: [0]=?, [1]=name, [2]=code, [3]=price
	if len(parts) < 4 {
		return 0, parts, errors.New("not enough fields")
	}
	rate, ok := anyToFloat(parts[3])
	if !ok {
		return 0, parts, errors.New("invalid price field")
	}
	return rate, parts, nil
}
