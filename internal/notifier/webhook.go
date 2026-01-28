package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"value-sniffer-radar/internal/config"
)

type Webhook struct {
	url     string
	headers map[string]string
	timeout time.Duration
}

func NewWebhook(c config.NotifierConfig) (*Webhook, error) {
	if c.URL == "" {
		return nil, fmt.Errorf("webhook.url required")
	}
	timeout := time.Duration(c.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Webhook{
		url:     c.URL,
		headers: c.Headers,
		timeout: timeout,
	}, nil
}

func (w *Webhook) Name() string { return "webhook" }

func (w *Webhook) Notify(ctx context.Context, events []Event) error {
	payload := map[string]any{
		"events": events,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range w.headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: w.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook status: %s", resp.Status)
	}
	return nil
}
