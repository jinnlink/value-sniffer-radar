package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// APIClient calls an OpenAI-compatible Chat Completions endpoint.
// It is intentionally minimal to remain provider-agnostic.
type APIClient struct {
	BaseURL   string
	APIKey    string
	Model     string
	Timeout   time.Duration
	UserAgent string
}

func (c APIClient) Name() string { return "api" }

type chatReq struct {
	Model       string    `json:"model"`
	Messages    []chatMsg `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type chatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResp struct {
	Choices []struct {
		Message chatMsg `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (c APIClient) Complete(ctx context.Context, prompt string) (string, error) {
	if strings.TrimSpace(c.APIKey) == "" {
		return "", fmt.Errorf("missing api key")
	}
	base := strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	model := strings.TrimSpace(c.Model)
	if model == "" {
		model = "gpt-4o-mini"
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	reqBody := chatReq{
		Model: model,
		Messages: []chatMsg{
			{Role: "system", Content: "You output strict JSON only."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.2,
		MaxTokens:   400,
	}
	b, _ := json.Marshal(reqBody)

	u := base + "/chat/completions"
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	hreq.Header.Set("Content-Type", "application/json")
	hreq.Header.Set("Authorization", "Bearer "+c.APIKey)
	if ua := strings.TrimSpace(c.UserAgent); ua != "" {
		hreq.Header.Set("User-Agent", ua)
	}

	hc := &http.Client{Timeout: timeout}
	resp, err := hc.Do(hreq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("api status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var cr chatResp
	if err := json.Unmarshal(body, &cr); err != nil {
		return "", err
	}
	if cr.Error != nil && cr.Error.Message != "" {
		return "", fmt.Errorf("api error: %s", cr.Error.Message)
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}
	return cr.Choices[0].Message.Content, nil
}
