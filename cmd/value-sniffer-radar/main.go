package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"value-sniffer-radar/internal/config"
	"value-sniffer-radar/internal/engine"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "Path to config YAML")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	e, err := engine.New(cfg)
	if err != nil {
		log.Fatalf("init engine: %v", err)
	}

	log.Printf("value-sniffer-radar started, interval=%s", time.Duration(cfg.Engine.IntervalSeconds)*time.Second)
	if cfg.Engine.TradeDateMode == "fixed" {
		log.Printf("trade_date fixed: %s", cfg.Engine.FixedTradeDate)
	} else {
		log.Printf("trade_date mode: %s", cfg.Engine.TradeDateMode)
	}

	if err := e.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
