package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Tushare    TushareConfig    `yaml:"tushare"`
	Engine     EngineConfig     `yaml:"engine"`
	Marketdata MarketdataConfig `yaml:"marketdata"`
	Notifiers  []NotifierConfig `yaml:"notifiers"`
	Signals    []SignalConfig   `yaml:"signals"`
}

type TushareConfig struct {
	BaseURL        string `yaml:"base_url"`
	TokenEnv       string `yaml:"token_env"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

type EngineConfig struct {
	IntervalSeconds int    `yaml:"interval_seconds"`
	TradeDateMode   string `yaml:"trade_date_mode"`  // latest_open | fixed
	FixedTradeDate  string `yaml:"fixed_trade_date"` // YYYYMMDD
	MaxAPIRetries   int    `yaml:"max_api_retries"`
	DedupeSeconds   int    `yaml:"dedupe_seconds"`     // default 3600; set -1 to disable
	MaxEventsPerRun int    `yaml:"max_events_per_run"` // default 50; 0 means no limit

	// Tier controls: "action" (high quality) vs "observe" (broad coverage).
	ActionMaxEventsPerRun  int `yaml:"action_max_events_per_run"`  // default 10; 0 means unlimited
	ObserveMaxEventsPerRun int `yaml:"observe_max_events_per_run"` // default 50; 0 means unlimited

	ActionSymbolCooldownSeconds  int `yaml:"action_symbol_cooldown_seconds"`  // default 1800; set -1 to disable
	ObserveSymbolCooldownSeconds int `yaml:"observe_symbol_cooldown_seconds"` // default 7200; set -1 to disable

	ActionMaxEventsPerDay  int `yaml:"action_max_events_per_day"`  // default 30; 0 means unlimited
	ObserveMaxEventsPerDay int `yaml:"observe_max_events_per_day"` // default 200; 0 means unlimited

	// Action quality gate: net advantage after costs (pct points).
	// If <=0, net-edge policy is disabled (backward compatible).
	ActionNetEdgeMinPct float64 `yaml:"action_net_edge_min_pct"`

	// Cost defaults (pct points). Individual signals can override by writing into event.data.
	DefaultSpreadPct   float64            `yaml:"default_spread_pct"`
	DefaultSlippagePct float64            `yaml:"default_slippage_pct"`
	DefaultFeePct      float64            `yaml:"default_fee_pct"`
	FeePctByMarket     map[string]float64 `yaml:"fee_pct_by_market"`

	// Per-signal action budgets (daily).
	// Example:
	//   action_max_events_per_signal_per_day:
	//     cn_repo_sniper_action: 10
	ActionMaxEventsPerSignalPerDay map[string]int `yaml:"action_max_events_per_signal_per_day"`
}

type MarketdataConfig struct {
	Enabled bool `yaml:"enabled"`

	TimeoutMS int `yaml:"timeout_ms"` // per-provider timeout

	// Fusion thresholds (defaults are conservative; tune via paper evaluation).
	RequiredSources int     `yaml:"required_sources"` // default 2
	MaxAbsDiff      float64 `yaml:"max_abs_diff"`     // default 0.05 (pct points)
	StalenessSec    int     `yaml:"staleness_sec"`    // default 10
	MinValid        float64 `yaml:"min_valid"`        // default 0
	MaxValid        float64 `yaml:"max_valid"`        // default 20

	// Circuit breaker
	FailThreshold    int `yaml:"fail_threshold"`    // default 3
	OutlierThreshold int `yaml:"outlier_threshold"` // default 3
	CooldownSec      int `yaml:"cooldown_sec"`      // default 120

	Providers []MarketdataProviderConfig `yaml:"providers"`
}

type MarketdataProviderConfig struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // eastmoney_repo | tencent_repo

	// eastmoney_repo
	BaseURL     string  `yaml:"base_url"`
	Fields      string  `yaml:"fields"`
	RateDivisor float64 `yaml:"rate_divisor"` // optional (default 1.0)

	// tencent_repo
	QuoteURL string `yaml:"quote_url"`
}

type NotifierConfig struct {
	Type string `yaml:"type"` // stdout | email | webhook | aival_queue | paper_log

	// email
	SMTPHost      string   `yaml:"smtp_host"`
	SMTPPort      int      `yaml:"smtp_port"`
	UsernameEnv   string   `yaml:"username_env"`
	PasswordEnv   string   `yaml:"password_env"`
	From          string   `yaml:"from"`
	To            []string `yaml:"to"`
	SubjectPrefix string   `yaml:"subject_prefix"`

	// webhook
	URL            string            `yaml:"url"`
	TimeoutSeconds int               `yaml:"timeout_seconds"`
	Headers        map[string]string `yaml:"headers"`

	// aival_queue (AI-Value / AstrBot file-queue)
	QueueDir string   `yaml:"queue_dir"`
	Market   string   `yaml:"market"`
	Tags     []string `yaml:"tags"`

	// paper_log (append JSONL for evaluation)
	FilePath string `yaml:"file_path"`
}

type SignalConfig struct {
	Type               string `yaml:"type"` // cb_premium | cb_double_low | fund_premium | cn_repo_sniper | cn_repo_realtime
	Name               string `yaml:"name"` // instance name (optional). Allows multiple entries of same type.
	Enabled            bool   `yaml:"enabled"`
	Tier               string `yaml:"tier"`                 // action | observe
	MinIntervalSeconds int    `yaml:"min_interval_seconds"` // 0 uses engine interval

	// shared
	MinAmount float64 `yaml:"min_amount"`
	TopN      int     `yaml:"top_n"`
	ConfirmK  int     `yaml:"confirm_k"` // optional k-confirm for action triggers (default 1)

	// cb_premium
	PremiumPctLow  float64 `yaml:"premium_pct_low"`
	PremiumPctHigh float64 `yaml:"premium_pct_high"`

	// cb_double_low
	MaxDoubleLow float64 `yaml:"max_double_low"`

	// fund_premium
	Market          string `yaml:"market"`
	PickTopByAmount int    `yaml:"pick_top_by_amount"`

	// cn_repo_sniper (reverse repo yield monitor)
	RepoCodes   []string `yaml:"repo_codes"`    // e.g. ["204001.SH","131810.SZ"]
	MinYieldPct float64  `yaml:"min_yield_pct"` // threshold on weighted rate (%)
	WindowStart string   `yaml:"window_start"`  // "HH:MM" or "HHMM" (optional)
	WindowEnd   string   `yaml:"window_end"`    // "HH:MM" or "HHMM" (optional)
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.normalizeAndValidate(filepath.Dir(path)); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) normalizeAndValidate(baseDir string) error {
	if c.Tushare.TokenEnv == "" {
		c.Tushare.TokenEnv = "TUSHARE_TOKEN"
	}
	if c.Tushare.BaseURL == "" {
		c.Tushare.BaseURL = "https://api.tushare.pro"
	}
	if c.Tushare.TimeoutSeconds <= 0 {
		c.Tushare.TimeoutSeconds = 20
	}
	if c.Engine.IntervalSeconds <= 0 {
		c.Engine.IntervalSeconds = 300
	}
	if c.Engine.TradeDateMode == "" {
		c.Engine.TradeDateMode = "latest_open"
	}
	if c.Engine.MaxAPIRetries <= 0 {
		c.Engine.MaxAPIRetries = 3
	}
	if c.Engine.DedupeSeconds == 0 {
		c.Engine.DedupeSeconds = 3600
	} else if c.Engine.DedupeSeconds < -1 {
		return errors.New("engine.dedupe_seconds must be -1 (disable) or >= 0")
	}
	if c.Engine.MaxEventsPerRun < 0 {
		return errors.New("engine.max_events_per_run must be >= 0")
	}
	if c.Engine.MaxEventsPerRun == 0 {
		c.Engine.MaxEventsPerRun = 50
	}
	if c.Engine.ActionMaxEventsPerRun < 0 || c.Engine.ObserveMaxEventsPerRun < 0 {
		return errors.New("engine.action_max_events_per_run / observe_max_events_per_run must be >= 0")
	}
	if c.Engine.ActionMaxEventsPerRun == 0 {
		c.Engine.ActionMaxEventsPerRun = 10
	}
	if c.Engine.ObserveMaxEventsPerRun == 0 {
		c.Engine.ObserveMaxEventsPerRun = 50
	}
	if c.Engine.ActionSymbolCooldownSeconds == 0 {
		c.Engine.ActionSymbolCooldownSeconds = 1800
	} else if c.Engine.ActionSymbolCooldownSeconds < -1 {
		return errors.New("engine.action_symbol_cooldown_seconds must be -1 (disable) or >= 0")
	}
	if c.Engine.ObserveSymbolCooldownSeconds == 0 {
		c.Engine.ObserveSymbolCooldownSeconds = 7200
	} else if c.Engine.ObserveSymbolCooldownSeconds < -1 {
		return errors.New("engine.observe_symbol_cooldown_seconds must be -1 (disable) or >= 0")
	}
	if c.Engine.ActionMaxEventsPerDay < 0 || c.Engine.ObserveMaxEventsPerDay < 0 {
		return errors.New("engine.action_max_events_per_day / observe_max_events_per_day must be >= 0")
	}
	if c.Engine.ActionMaxEventsPerDay == 0 {
		c.Engine.ActionMaxEventsPerDay = 30
	}
	if c.Engine.ObserveMaxEventsPerDay == 0 {
		c.Engine.ObserveMaxEventsPerDay = 200
	}
	if c.Engine.DefaultSpreadPct < 0 || c.Engine.DefaultSlippagePct < 0 || c.Engine.DefaultFeePct < 0 {
		return errors.New("engine.default_spread_pct / default_slippage_pct / default_fee_pct must be >= 0")
	}
	if c.Engine.FeePctByMarket == nil {
		c.Engine.FeePctByMarket = map[string]float64{}
	}
	if c.Engine.ActionMaxEventsPerSignalPerDay == nil {
		c.Engine.ActionMaxEventsPerSignalPerDay = map[string]int{}
	}
	for k, v := range c.Engine.ActionMaxEventsPerSignalPerDay {
		if v < 0 {
			return errors.New("engine.action_max_events_per_signal_per_day values must be >= 0")
		}
		if k == "" {
			delete(c.Engine.ActionMaxEventsPerSignalPerDay, k)
		}
	}

	// marketdata defaults (optional)
	if c.Marketdata.TimeoutMS <= 0 {
		c.Marketdata.TimeoutMS = 1500
	}
	if c.Marketdata.RequiredSources <= 0 {
		c.Marketdata.RequiredSources = 2
	}
	if c.Marketdata.MaxAbsDiff <= 0 {
		c.Marketdata.MaxAbsDiff = 0.05
	}
	if c.Marketdata.StalenessSec <= 0 {
		c.Marketdata.StalenessSec = 10
	}
	if c.Marketdata.MinValid == 0 {
		c.Marketdata.MinValid = 0
	}
	if c.Marketdata.MaxValid == 0 {
		c.Marketdata.MaxValid = 20
	}
	if c.Marketdata.FailThreshold <= 0 {
		c.Marketdata.FailThreshold = 3
	}
	if c.Marketdata.OutlierThreshold <= 0 {
		c.Marketdata.OutlierThreshold = 3
	}
	if c.Marketdata.CooldownSec <= 0 {
		c.Marketdata.CooldownSec = 120
	}
	for i := range c.Marketdata.Providers {
		p := &c.Marketdata.Providers[i]
		if p.RateDivisor == 0 {
			p.RateDivisor = 1.0
		}
	}
	if c.Engine.TradeDateMode != "latest_open" && c.Engine.TradeDateMode != "fixed" {
		return errors.New("engine.trade_date_mode must be latest_open or fixed")
	}
	if c.Engine.TradeDateMode == "fixed" && c.Engine.FixedTradeDate == "" {
		return errors.New("engine.fixed_trade_date required when trade_date_mode=fixed")
	}
	if c.RequiresTushare() && os.Getenv(c.Tushare.TokenEnv) == "" {
		return errors.New("missing Tushare token env: " + c.Tushare.TokenEnv)
	}
	if len(c.Notifiers) == 0 {
		return errors.New("at least one notifier required (e.g. stdout)")
	}
	for i := range c.Notifiers {
		n := &c.Notifiers[i]
		if n.Type == "paper_log" && n.FilePath != "" && !filepath.IsAbs(n.FilePath) {
			n.FilePath = filepath.Join(baseDir, n.FilePath)
		}
	}
	for i := range c.Signals {
		s := &c.Signals[i]
		if s.Tier == "" {
			s.Tier = "action"
		}
		if s.Tier != "action" && s.Tier != "observe" {
			return errors.New("signals[].tier must be action or observe")
		}
		if s.MinIntervalSeconds < 0 {
			return errors.New("signals[].min_interval_seconds must be >= 0")
		}
		if s.ConfirmK < 0 {
			return errors.New("signals[].confirm_k must be >= 0")
		}
		if s.ConfirmK == 0 {
			s.ConfirmK = 1
		}
	}
	return nil
}

func (c *Config) RequiresTushare() bool {
	// Trade date resolution via trade_cal needs Tushare.
	if c.Engine.TradeDateMode == "latest_open" {
		return true
	}
	// Signal data needs Tushare for these types.
	for _, s := range c.Signals {
		if !s.Enabled {
			continue
		}
		switch s.Type {
		case "cb_premium", "cb_double_low", "fund_premium", "cn_repo_sniper":
			return true
		}
	}
	return false
}
