package optimizer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// RepoLabel matches one line in labels.repo.jsonl produced by cmd/value-sniffer-radar-labeler.
// Keep this struct local to optimizer to avoid import cycles (labeler imports optimizer).
type RepoLabel struct {
	EventID   string `json:"event_id"`
	Source    string `json:"source"`
	Symbol    string `json:"symbol"`
	TradeDate string `json:"trade_date"`

	WindowSec int `json:"window_sec"`
	Reward    int `json:"reward"`

	Confidence string `json:"confidence"`
	Reason     string `json:"reason"`
}

type LabelsIndex struct {
	ByEvent       map[string]map[int]RepoLabel // event_id -> window_sec -> label
	Windows       []int                        // sorted unique windows
	CountByWindow map[int]int                  // window_sec -> labels count
	TotalLabels   int
	Warnings      []string
}

func ReadLabelsJSONL(r io.Reader) (LabelsIndex, error) {
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 4*1024*1024)

	out := LabelsIndex{
		ByEvent:       map[string]map[int]RepoLabel{},
		CountByWindow: map[int]int{},
	}

	lineNo := 0
	for sc.Scan() {
		lineNo++
		s := strings.TrimSpace(sc.Text())
		s = strings.TrimPrefix(s, "\ufeff")
		if s == "" {
			continue
		}
		var l RepoLabel
		if err := json.Unmarshal([]byte(s), &l); err != nil {
			out.Warnings = append(out.Warnings, fmt.Sprintf("invalid_label_json line=%d", lineNo))
			continue
		}
		if l.EventID == "" || l.WindowSec <= 0 {
			out.Warnings = append(out.Warnings, fmt.Sprintf("invalid_label_fields line=%d", lineNo))
			continue
		}
		m, ok := out.ByEvent[l.EventID]
		if !ok {
			m = map[int]RepoLabel{}
			out.ByEvent[l.EventID] = m
		}
		// Append-only file; if duplicates exist, last one wins.
		m[l.WindowSec] = l
		out.CountByWindow[l.WindowSec]++
		out.TotalLabels++
	}
	if err := sc.Err(); err != nil {
		return out, err
	}

	for w := range out.CountByWindow {
		out.Windows = append(out.Windows, w)
	}
	sort.Ints(out.Windows)
	return out, nil
}

func (li LabelsIndex) Has(eventID string, windowSec int) bool {
	m, ok := li.ByEvent[eventID]
	if !ok {
		return false
	}
	_, ok = m[windowSec]
	return ok
}

func (li LabelsIndex) Get(eventID string, windowSec int) (RepoLabel, bool) {
	m, ok := li.ByEvent[eventID]
	if !ok {
		return RepoLabel{}, false
	}
	l, ok := m[windowSec]
	return l, ok
}

// DefaultPrimaryWindowSec picks the window with the most labels.
// If there are no labels, it returns 0.
func (li LabelsIndex) DefaultPrimaryWindowSec() int {
	bestW := 0
	bestN := 0
	for w, n := range li.CountByWindow {
		if n > bestN {
			bestN = n
			bestW = w
		}
	}
	return bestW
}

func ResolveReward(pr PaperRow, labels LabelsIndex, primaryWindowSec int) (reward int, ok bool, source string) {
	if primaryWindowSec > 0 {
		if l, ok := labels.Get(EventID(pr), primaryWindowSec); ok {
			return clampReward(l.Reward), true, "labels"
		}
	}
	if r, ok := RewardFromRow(pr); ok {
		return clampReward(r), true, "paper"
	}
	return 0, false, ""
}

func clampReward(v int) int {
	if v > 0 {
		return 1
	}
	return 0
}
