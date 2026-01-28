package tushare

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Options struct {
	BaseURL        string
	Token          string
	TimeoutSeconds int
	MaxRetries     int
}

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	maxRetries int
}

func New(opt Options) *Client {
	timeout := time.Duration(opt.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	retries := opt.MaxRetries
	if retries <= 0 {
		retries = 3
	}
	return &Client{
		baseURL: opt.BaseURL,
		token:   opt.Token,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxRetries: retries,
	}
}

type request struct {
	APIName string         `json:"api_name"`
	Token   string         `json:"token"`
	Params  map[string]any `json:"params"`
	Fields  string         `json:"fields,omitempty"`
}

type response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Fields []string        `json:"fields"`
		Items  [][]interface{} `json:"items"`
	} `json:"data"`
}

func (c *Client) Query(ctx context.Context, apiName string, params map[string]any, fields []string) ([]map[string]interface{}, error) {
	reqBody := request{
		APIName: apiName,
		Token:   c.token,
		Params:  params,
	}
	if len(fields) > 0 {
		reqBody.Fields = joinFields(fields)
	}

	var lastErr error
	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		out, err := c.doOnce(ctx, reqBody)
		if err == nil {
			return out, nil
		}
		lastErr = err
		time.Sleep(time.Duration(attempt) * 400 * time.Millisecond)
	}
	return nil, lastErr
}

func (c *Client) doOnce(ctx context.Context, reqBody request) ([]map[string]interface{}, error) {
	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tushare http status=%s body=%s", resp.Status, string(raw))
	}

	var tr response
	if err := json.Unmarshal(raw, &tr); err != nil {
		return nil, fmt.Errorf("tushare parse: %w body=%s", err, string(raw))
	}
	if tr.Code != 0 {
		return nil, fmt.Errorf("tushare api error code=%d msg=%s", tr.Code, tr.Msg)
	}
	return rowsToMaps(tr.Data.Fields, tr.Data.Items), nil
}

func (c *Client) LatestOpenTradeDate(ctx context.Context, lookbackDays int) (string, error) {
	if lookbackDays <= 0 {
		lookbackDays = 30
	}
	end := time.Now()
	start := end.AddDate(0, 0, -lookbackDays)
	startStr := start.Format("20060102")
	endStr := end.Format("20060102")

	rows, err := c.Query(ctx, "trade_cal", map[string]any{
		"exchange":   "",
		"is_open":    "1",
		"start_date": startStr,
		"end_date":   endStr,
	}, []string{"cal_date", "is_open"})
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", errors.New("no open trade_date found from trade_cal")
	}
	latest := ""
	for _, r := range rows {
		d := GetString(r, "cal_date")
		if d > latest {
			latest = d
		}
	}
	if latest == "" {
		return "", errors.New("trade_cal returned empty cal_date")
	}
	return latest, nil
}

func rowsToMaps(fields []string, items [][]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	for _, row := range items {
		m := make(map[string]interface{}, len(fields))
		for i, f := range fields {
			if i < len(row) {
				m[f] = row[i]
			}
		}
		out = append(out, m)
	}
	return out
}

func joinFields(fields []string) string {
	if len(fields) == 0 {
		return ""
	}
	s := fields[0]
	for i := 1; i < len(fields); i++ {
		s += "," + fields[i]
	}
	return s
}

func GetString(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		// some numeric codes can be safely rendered without decimals
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func GetFloat(m map[string]interface{}, key string) float64 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return t
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return f
	default:
		return 0
	}
}

