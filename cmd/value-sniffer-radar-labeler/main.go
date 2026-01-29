package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"value-sniffer-radar/internal/config"
	"value-sniffer-radar/internal/labeler"
	"value-sniffer-radar/internal/marketdata"
)

type mockFusion struct {
	rate float64
	conf marketdata.Confidence
}

func (m mockFusion) FetchFusion(ctx context.Context, symbol string) (marketdata.FusionSnapshot, error) {
	_ = ctx
	return marketdata.FusionSnapshot{
		Symbol:           symbol,
		TS:               time.Now(),
		ConsensusRatePct: m.rate,
		Confidence:       m.conf,
		Reason:           "mock",
	}, nil
}

func main() {
	var configPath string
	var inPath string
	var outPath string
	var windows string
	var grace string
	var maxPerRun int
	var mockRate float64
	var mockConfidence string

	flag.StringVar(&configPath, "config", "config.yaml", "Path to config YAML")
	flag.StringVar(&inPath, "in", "", "Input paper_log JSONL path")
	flag.StringVar(&outPath, "out", "", "Output labels JSONL path")
	flag.StringVar(&windows, "windows", "10s,30s,5m", "Comma-separated windows (e.g. 10s,30s,5m)")
	flag.StringVar(&grace, "grace", "30s", "Grace period for late sampling (e.g. 30s)")
	flag.IntVar(&maxPerRun, "max", 200, "Max labels to write per run")
	flag.Float64Var(&mockRate, "mock-rate", math.NaN(), "Optional: use a mock fusion rate (no network). Example: 1.6")
	flag.StringVar(&mockConfidence, "mock-confidence", "PASS", "Mock confidence: PASS|FAIL (used only when -mock-rate is set)")
	flag.Parse()

	if inPath == "" {
		fmt.Fprintln(os.Stderr, "[error] missing -in")
		os.Exit(2)
	}
	if outPath == "" {
		outPath = ".\\state\\labels.repo.jsonl"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error] load config:", err.Error())
		os.Exit(1)
	}
	var md marketdata.Fusion
	if !math.IsNaN(mockRate) {
		conf := marketdata.ConfidencePass
		if strings.EqualFold(strings.TrimSpace(mockConfidence), "FAIL") {
			conf = marketdata.ConfidenceFail
		}
		md = mockFusion{rate: mockRate, conf: conf}
	} else {
		md, err = marketdata.Build(cfg.Marketdata)
		if err != nil {
			fmt.Fprintln(os.Stderr, "[error] init marketdata:", err.Error())
			os.Exit(1)
		}
	}

	ws, err := labeler.ParseWindows(windows)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error] parse windows:", err.Error())
		os.Exit(2)
	}
	gd, err := time.ParseDuration(grace)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error] parse grace:", err.Error())
		os.Exit(2)
	}

	lcfg := labeler.DefaultConfig()
	lcfg.Windows = ws
	lcfg.Grace = gd
	lcfg.MaxPerRun = maxPerRun

	r := labeler.New(cfg, md, lcfg)
	wrote, skipped, err := r.RunOnce(context.Background(), inPath, outPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error] run:", err.Error())
		os.Exit(1)
	}
	fmt.Printf("labels_written=%d skipped=%d out=%s\n", wrote, skipped, outPath)
}
