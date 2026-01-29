package optimizer

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type PaperRow struct {
	TS    string        `json:"ts"`
	Event PaperLogEvent `json:"event"`
}

type PaperLogEvent struct {
	Source    string            `json:"source"`
	TradeDate string            `json:"trade_date"`
	Market    string            `json:"market"`
	Symbol    string            `json:"symbol"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	Tags      map[string]string `json:"tags"`
	Data      map[string]any    `json:"data"`
}

func ReadJSONL(r io.Reader) ([]PaperRow, []string, error) {
	sc := bufio.NewScanner(r)
	// Allow long lines (raw providers etc.)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 4*1024*1024)

	var rows []PaperRow
	var warns []string
	lineNo := 0
	for sc.Scan() {
		lineNo++
		s := strings.TrimSpace(sc.Text())
		s = strings.TrimPrefix(s, "\ufeff")
		if s == "" {
			continue
		}
		var pr PaperRow
		if err := json.Unmarshal([]byte(s), &pr); err != nil {
			warns = append(warns, fmt.Sprintf("invalid_json line=%d", lineNo))
			continue
		}
		if pr.Event.Source == "" {
			warns = append(warns, fmt.Sprintf("missing_event_source line=%d", lineNo))
			continue
		}
		rows = append(rows, pr)
	}
	if err := sc.Err(); err != nil {
		return rows, warns, err
	}
	return rows, warns, nil
}

func RewardFromRow(pr PaperRow) (int, bool) {
	if pr.Event.Data == nil {
		return 0, false
	}
	v, ok := pr.Event.Data["reward"]
	if !ok {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		if t > 0 {
			return 1, true
		}
		return 0, true
	case int:
		if t != 0 {
			return 1, true
		}
		return 0, true
	case bool:
		if t {
			return 1, true
		}
		return 0, true
	case string:
		tt := strings.TrimSpace(strings.ToLower(t))
		if tt == "1" || tt == "true" || tt == "yes" || tt == "y" {
			return 1, true
		}
		if tt == "0" || tt == "false" || tt == "no" || tt == "n" {
			return 0, true
		}
		return 0, false
	default:
		return 0, false
	}
}

func EventID(pr PaperRow) string {
	h := sha256.Sum256([]byte(pr.TS + "|" + pr.Event.Source + "|" + pr.Event.Symbol + "|" + pr.Event.TradeDate + "|" + pr.Event.Title))
	return hex.EncodeToString(h[:])
}
