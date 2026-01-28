package notifier

import (
	"context"
	"fmt"
	"strings"
)

type Stdout struct{}

func NewStdout() *Stdout { return &Stdout{} }

func (s *Stdout) Name() string { return "stdout" }

func (s *Stdout) Notify(_ context.Context, events []Event) error {
	for _, e := range events {
		fmt.Println(formatEvent(e))
		fmt.Println(strings.Repeat("-", 60))
	}
	return nil
}

func formatEvent(e Event) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("[%s] %s\n", e.Source, e.Title))
	if e.Tags != nil {
		if t := strings.TrimSpace(e.Tags["tier"]); t != "" {
			b.WriteString(fmt.Sprintf("tier: %s\n", t))
		}
	}
	if e.Symbol != "" {
		b.WriteString(fmt.Sprintf("symbol: %s\n", e.Symbol))
	}
	if e.TradeDate != "" {
		b.WriteString(fmt.Sprintf("trade_date: %s\n", e.TradeDate))
	}
	if e.Body != "" {
		b.WriteString(e.Body)
		if !strings.HasSuffix(e.Body, "\n") {
			b.WriteString("\n")
		}
	}
	return b.String()
}
