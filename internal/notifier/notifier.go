package notifier

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"value-sniffer-radar/internal/config"
)

type Event struct {
	Source    string                 `json:"source"`
	TradeDate string                 `json:"trade_date"`
	Market    string                 `json:"market,omitempty"`
	Symbol    string                 `json:"symbol,omitempty"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Tags      map[string]string      `json:"tags,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

type Notifier interface {
	Name() string
	Notify(ctx context.Context, events []Event) error
}

func BuildAll(cfgs []config.NotifierConfig) ([]Notifier, error) {
	var out []Notifier
	for _, c := range cfgs {
		switch c.Type {
		case "stdout":
			out = append(out, NewStdout())
		case "email":
			n, err := NewEmail(c)
			if err != nil {
				return nil, err
			}
			out = append(out, n)
		case "webhook":
			n, err := NewWebhook(c)
			if err != nil {
				return nil, err
			}
			out = append(out, n)
		case "aival_queue":
			n, err := NewAivalQueue(c)
			if err != nil {
				return nil, err
			}
			out = append(out, n)
		case "paper_log":
			n, err := NewPaperLog(c)
			if err != nil {
				return nil, err
			}
			out = append(out, n)
		default:
			return nil, fmt.Errorf("unknown notifier type: %s", c.Type)
		}
	}
	if len(out) == 0 {
		return nil, errors.New("no notifiers configured")
	}
	return out, nil
}

func JSON(events []Event) string {
	b, _ := json.MarshalIndent(events, "", "  ")
	return string(b)
}
