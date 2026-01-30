package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"time"

	"value-sniffer-radar/internal/optimizer"
)

func main() {
	var inPath string
	var labelsPath string
	var labelWindowSec int
	var outMD string
	var outReco string
	var seed int64
	var slots int

	flag.StringVar(&inPath, "in", "", "Input JSONL path (paper_log). Use '-' for stdin.")
	flag.StringVar(&labelsPath, "labels", "", "Optional labels.repo.jsonl path (append-only).")
	flag.IntVar(&labelWindowSec, "label-window-sec", 0, "Which labels window (sec) to use for bandit updates. 0=auto.")
	flag.StringVar(&outMD, "out-md", "", "Optional Markdown output path.")
	flag.StringVar(&outReco, "out-reco", "", "Optional reco JSON path (e.g. .\\state\\optimizer.reco.json).")
	flag.Int64Var(&seed, "seed", 7, "RNG seed for deterministic suggestions.")
	flag.IntVar(&slots, "slots", 10, "How many action slots to suggest.")
	flag.Parse()

	if inPath == "" {
		fmt.Fprintln(os.Stderr, "[error] missing -in")
		os.Exit(2)
	}

	var (
		rows  []optimizer.PaperRow
		warns []string
		err   error
	)

	label := inPath
	if inPath == "-" {
		rows, warns, err = optimizer.ReadJSONL(os.Stdin)
		label = "<stdin>"
	} else {
		f, e := os.Open(inPath)
		if e != nil {
			fmt.Fprintln(os.Stderr, "[error] open:", e.Error())
			os.Exit(2)
		}
		defer f.Close()
		rows, warns, err = optimizer.ReadJSONL(f)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error] read jsonl:", err.Error())
		os.Exit(1)
	}

	var labels optimizer.LabelsIndex
	if labelsPath != "" {
		f, e := os.Open(labelsPath)
		if e != nil {
			if os.IsNotExist(e) {
				labels = optimizer.LabelsIndex{
					ByEvent:       map[string]map[int]optimizer.RepoLabel{},
					CountByWindow: map[int]int{},
					Warnings:      []string{"labels_not_found"},
				}
			} else {
				fmt.Fprintln(os.Stderr, "[error] open labels:", e.Error())
				os.Exit(2)
			}
		} else {
			defer f.Close()
			labels, err = optimizer.ReadLabelsJSONL(f)
			if err != nil {
				fmt.Fprintln(os.Stderr, "[error] read labels:", err.Error())
				os.Exit(1)
			}
		}
	}

	primaryWindowSec := labelWindowSec
	if primaryWindowSec <= 0 && labelsPath != "" {
		primaryWindowSec = labels.DefaultPrimaryWindowSec()
	}

	// Unique event ids in the input (coverage denominator).
	inputEventIDs := map[string]string{} // event_id -> source
	for _, pr := range rows {
		id := optimizer.EventID(pr)
		if _, ok := inputEventIDs[id]; ok {
			continue
		}
		inputEventIDs[id] = pr.Event.Source
	}

	var coverage []optimizer.CoverageStat
	if labelsPath != "" && len(labels.Windows) > 0 && len(inputEventIDs) > 0 {
		total := len(inputEventIDs)
		for _, w := range labels.Windows {
			labeled := 0
			for id := range inputEventIDs {
				if labels.Has(id, w) {
					labeled++
				}
			}
			pct := 0.0
			if total > 0 {
				pct = float64(labeled) * 100 / float64(total)
			}
			coverage = append(coverage, optimizer.CoverageStat{
				WindowSec:     w,
				TotalEvents:   total,
				LabeledEvents: labeled,
				CoveragePct:   pct,
			})
		}
	}

	// Reward rates by signal/window (from labels, filtered to input event ids).
	type rrKey struct {
		signal string
		window int
	}
	type rrAgg struct {
		n   int
		sum int
	}
	rr := map[rrKey]*rrAgg{}
	if labelsPath != "" && len(labels.ByEvent) > 0 && len(labels.Windows) > 0 && len(inputEventIDs) > 0 {
		for id, srcFallback := range inputEventIDs {
			for _, w := range labels.Windows {
				l, ok := labels.Get(id, w)
				if !ok {
					continue
				}
				sig := l.Source
				if sig == "" {
					sig = srcFallback
				}
				k := rrKey{signal: sig, window: w}
				a := rr[k]
				if a == nil {
					a = &rrAgg{}
					rr[k] = a
				}
				a.n++
				if l.Reward > 0 {
					a.sum++
				}
			}
		}
	}
	var rewardRates []optimizer.RewardRateStat
	for k, a := range rr {
		pct := 0.0
		if a.n > 0 {
			pct = float64(a.sum) * 100 / float64(a.n)
		}
		rewardRates = append(rewardRates, optimizer.RewardRateStat{
			Signal:    k.signal,
			WindowSec: k.window,
			N:         a.n,
			RewardSum: a.sum,
			RatePct:   pct,
		})
	}
	// Deterministic render ordering.
	// (Signal asc, window asc)
	if len(rewardRates) > 1 {
		sort.Slice(rewardRates, func(i, j int) bool {
			if rewardRates[i].Signal == rewardRates[j].Signal {
				return rewardRates[i].WindowSec < rewardRates[j].WindowSec
			}
			return rewardRates[i].Signal < rewardRates[j].Signal
		})
	}

	b := optimizer.NewBandit()
	withReward := 0
	fromLabels := 0
	fromPaper := 0
	for _, pr := range rows {
		key := pr.Event.Source
		reward, ok, src := optimizer.ResolveReward(pr, labels, primaryWindowSec)
		if ok {
			withReward++
			b.Update(key, reward)
			if src == "labels" {
				fromLabels++
			} else if src == "paper" {
				fromPaper++
			}
		} else {
			// Still register arm so it appears in the report.
			b.Ensure(key)
		}
	}

	rng := rand.New(rand.NewSource(seed))
	alloc, _ := b.SuggestAllocation(rng, slots)
	quotaReco := optimizer.SuggestQuotas(rand.New(rand.NewSource(seed)), b, slots)
	rep := optimizer.Report{
		GeneratedAt:       time.Now(),
		InputPath:         label,
		LabelsPath:        labelsPath,
		PrimaryWindowSec:  primaryWindowSec,
		Warnings:          warns,
		LabelWarnings:     labels.Warnings,
		UniqueEvents:      len(inputEventIDs),
		RewardsUsed:       withReward,
		RewardsFromLabels: fromLabels,
		RewardsFromPaper:  fromPaper,
		Coverage:          coverage,
		RewardRates:       rewardRates,
		ArmsTotal:         len(b.Arms),
		Alloc:             alloc,
	}
	md := optimizer.RenderMarkdown(rep)

	fmt.Print(md)

	if outMD != "" {
		if err := os.MkdirAll(filepath.Dir(outMD), 0o755); err == nil {
			_ = os.WriteFile(outMD, []byte(md), 0o644)
		}
	}

	if outReco != "" {
		reco := optimizer.BuildRecommendation(time.Now(), label, labelsPath, primaryWindowSec, slots, quotaReco, b)
		_ = optimizer.WriteReco(outReco, reco)
	}

	_ = withReward
}
