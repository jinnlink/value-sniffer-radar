package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"value-sniffer-radar/internal/optimizer"
)

func main() {
	var inPath string
	var outMD string
	var seed int64
	var slots int

	flag.StringVar(&inPath, "in", "", "Input JSONL path (paper_log). Use '-' for stdin.")
	flag.StringVar(&outMD, "out-md", "", "Optional Markdown output path.")
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

	b := optimizer.NewBandit()
	withReward := 0
	for _, pr := range rows {
		key := pr.Event.Source
		reward, ok := optimizer.RewardFromRow(pr)
		if ok {
			withReward++
			b.Update(key, reward)
		} else {
			// Still register arm so it appears in the report.
			b.Ensure(key)
		}
	}

	rng := rand.New(rand.NewSource(seed))
	alloc, _ := b.SuggestAllocation(rng, slots)
	rep := optimizer.Report{
		GeneratedAt: time.Now(),
		InputPath:   label,
		Warnings:    warns,
		ArmsTotal:   len(b.Arms),
		Alloc:       alloc,
	}
	md := optimizer.RenderMarkdown(rep)

	fmt.Print(md)

	if outMD != "" {
		if err := os.MkdirAll(filepath.Dir(outMD), 0o755); err == nil {
			_ = os.WriteFile(outMD, []byte(md), 0o644)
		}
	}

	_ = withReward
}

