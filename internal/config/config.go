package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/benvon/whats-flying-over-me/internal/cataloger"
)

// Config holds the application configuration.
type Config struct {
	ScrapeInterval time.Duration
	RadiusKm       float64
	AltitudeMax    int
	BaseLat        float64
	BaseLon        float64
	DataURL        string
	Notifier       NotifierConfig
	AlertDedupe    AlertDedupeConfig
	Cataloger      cataloger.ElasticSearchConfig
}

// Duration is a custom type that can unmarshal from string
type Duration time.Duration

func (d *Duration) UnmarshalJSON(data []byte) error {
	// Remove quotes from the string
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// Parse the duration
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(duration)
	return nil
}

// ConfigJSON is used for JSON unmarshaling
type ConfigJSON struct {
	ScrapeInterval Duration `json:"ScrapeInterval"`
	RadiusKm       float64  `json:"RadiusKm"`
	AltitudeMax    int      `json:"AltitudeMax"`
	BaseLat        float64  `json:"BaseLat"`
	BaseLon        float64  `json:"BaseLon"`
	DataURL        string   `json:"DataURL"`
	Notifier       struct {
		Webhook struct {
			Enabled bool     `json:"Enabled"`
			URL     string   `json:"URL"`
			Timeout Duration `json:"Timeout"`
		} `json:"Webhook"`
		RabbitMQ struct {
			Enabled    bool     `json:"Enabled"`
			URL        string   `json:"URL"`
			Exchange   string   `json:"Exchange"`
			RoutingKey string   `json:"RoutingKey"`
			Timeout    Duration `json:"Timeout"`
		} `json:"RabbitMQ"`
		Console bool `json:"Console"`
	} `json:"Notifier"`
	AlertDedupe struct {
		Enabled     bool     `json:"Enabled"`
		BlockoutMin Duration `json:"BlockoutMin"`
	} `json:"AlertDedupe"`
	Cataloger struct {
		Enabled    bool     `json:"Enabled"`
		URL        string   `json:"URL"`
		Index      string   `json:"Index"`
		Username   string   `json:"Username"`
		Password   string   `json:"Password"`
		Timeout    Duration `json:"Timeout"`
		MaxRetries int      `json:"MaxRetries"`
	} `json:"Cataloger"`
}

// UnmarshalJSON implements custom JSON unmarshaling for Config
func (c *Config) UnmarshalJSON(data []byte) error {
	var configJSON ConfigJSON
	if err := json.Unmarshal(data, &configJSON); err != nil {
		return err
	}

	// Convert Duration types to time.Duration
	c.ScrapeInterval = time.Duration(configJSON.ScrapeInterval)
	c.RadiusKm = configJSON.RadiusKm
	c.AltitudeMax = configJSON.AltitudeMax
	c.BaseLat = configJSON.BaseLat
	c.BaseLon = configJSON.BaseLon
	c.DataURL = configJSON.DataURL

	// Copy Notifier fields
	c.Notifier.Webhook.Enabled = configJSON.Notifier.Webhook.Enabled
	c.Notifier.Webhook.URL = configJSON.Notifier.Webhook.URL
	c.Notifier.Webhook.Timeout = time.Duration(configJSON.Notifier.Webhook.Timeout)
	c.Notifier.RabbitMQ.Enabled = configJSON.Notifier.RabbitMQ.Enabled
	c.Notifier.RabbitMQ.URL = configJSON.Notifier.RabbitMQ.URL
	c.Notifier.RabbitMQ.Exchange = configJSON.Notifier.RabbitMQ.Exchange
	c.Notifier.RabbitMQ.RoutingKey = configJSON.Notifier.RabbitMQ.RoutingKey
	c.Notifier.RabbitMQ.Timeout = time.Duration(configJSON.Notifier.RabbitMQ.Timeout)
	c.Notifier.Console = configJSON.Notifier.Console

	// Copy AlertDedupe fields
	c.AlertDedupe.Enabled = configJSON.AlertDedupe.Enabled
	c.AlertDedupe.BlockoutMin = time.Duration(configJSON.AlertDedupe.BlockoutMin)

	// Copy Cataloger fields
	c.Cataloger.Enabled = configJSON.Cataloger.Enabled
	c.Cataloger.URL = configJSON.Cataloger.URL
	c.Cataloger.Index = configJSON.Cataloger.Index
	c.Cataloger.Username = configJSON.Cataloger.Username
	c.Cataloger.Password = configJSON.Cataloger.Password
	c.Cataloger.Timeout = time.Duration(configJSON.Cataloger.Timeout)
	c.Cataloger.MaxRetries = configJSON.Cataloger.MaxRetries

	return nil
}

// NotifierConfig holds notifier settings.
type NotifierConfig struct {
	Webhook  WebhookConfig
	RabbitMQ RabbitMQConfig
	Console  bool
}

// WebhookConfig holds webhook notifier settings.
type WebhookConfig struct {
	Enabled bool
	URL     string
	Timeout time.Duration
}

// RabbitMQConfig holds RabbitMQ notifier settings.
type RabbitMQConfig struct {
	Enabled    bool
	URL        string
	Exchange   string
	RoutingKey string
	Timeout    time.Duration
}

// AlertDedupeConfig holds alert deduplication settings.
type AlertDedupeConfig struct {
	Enabled     bool
	BlockoutMin time.Duration
}

const (
	envConfigFile = "WFO_CONFIG"
	envInterval   = "WFO_INTERVAL"
	envRadius     = "WFO_RADIUS_KM"
	envAltitude   = "WFO_ALTITUDE_MAX"
	envBaseLat    = "WFO_BASE_LAT"
	envBaseLon    = "WFO_BASE_LON"
	envDataURL    = "WFO_DATA_URL"

	// Webhook settings
	envWebhookEnabled = "WFO_WEBHOOK_ENABLED"
	envWebhookURL     = "WFO_WEBHOOK_URL"
	envWebhookTimeout = "WFO_WEBHOOK_TIMEOUT"

	// RabbitMQ settings
	envRabbitMQEnabled    = "WFO_RABBITMQ_ENABLED"
	envRabbitMQURL        = "WFO_RABBITMQ_URL"
	envRabbitMQExchange   = "WFO_RABBITMQ_EXCHANGE"
	envRabbitMQRoutingKey = "WFO_RABBITMQ_ROUTING_KEY"
	envRabbitMQTimeout    = "WFO_RABBITMQ_TIMEOUT"

	// Alert deduplication settings
	envAlertDedupeEnabled = "WFO_ALERT_DEDUPE_ENABLED"
	envAlertBlockoutMin   = "WFO_ALERT_BLOCKOUT_MIN"

	// Cataloging settings
	envCatalogerEnabled    = "WFO_CATALOGER_ENABLED"
	envCatalogerURL        = "WFO_CATALOGER_URL"
	envCatalogerIndex      = "WFO_CATALOGER_INDEX"
	envCatalogerUsername   = "WFO_CATALOGER_USERNAME"
	envCatalogerPassword   = "WFO_CATALOGER_PASSWORD" // #nosec G101 -- this is a test password
	envCatalogerTimeout    = "WFO_CATALOGER_TIMEOUT"
	envCatalogerMaxRetries = "WFO_CATALOGER_MAX_RETRIES"
)

// Load reads configuration from config file, environment variables and command line flags
// in order of increasing precedence.
func Load() Config {
	return LoadWithFlagSet(flag.CommandLine)
}

// LoadWithFlagSet reads configuration using a specific flag set (useful for testing)
func LoadWithFlagSet(flagSet *flag.FlagSet) Config {
	// Defaults
	cfg := Config{
		ScrapeInterval: time.Minute,
		RadiusKm:       25.0,
		AltitudeMax:    10000,
		DataURL:        "http://localhost:8080/data/aircraft.json",
		Notifier: NotifierConfig{
			Console: true, // Default to console logging only
		},
		AlertDedupe: AlertDedupeConfig{
			Enabled:     true,
			BlockoutMin: 15 * time.Minute,
		},
		Cataloger: cataloger.ElasticSearchConfig{
			Enabled:    false, // Default to disabled
			Index:      "aircraft",
			Timeout:    30 * time.Second,
			MaxRetries: 3,
		},
	}

	// Define and parse command line flags
	flags := defineFlagsWithFlagSet(flagSet)
	// Parse flags - in test environment, os.Args might be set by the test
	if len(os.Args) > 1 {
		if err := flagSet.Parse(os.Args[1:]); err != nil {
			// Log error but continue with defaults
			_ = err // Suppress unused variable warning
		}
	}
	setFlags := getSetFlagsFromFlagSet(flagSet)

	// Load from config file
	loadFromConfigFile(&cfg, flags.configPath)

	// Load from environment variables
	loadFromEnvironment(&cfg)

	// Apply command line overrides
	applyCommandLineOverrides(&cfg, flags, setFlags)

	return cfg
}

// LoadWithFlagSetAndArgs reads configuration using a specific flag set and arguments (useful for testing)
func LoadWithFlagSetAndArgs(flagSet *flag.FlagSet, args []string) Config {
	// Defaults
	cfg := Config{
		ScrapeInterval: time.Minute,
		RadiusKm:       25.0,
		AltitudeMax:    10000,
		DataURL:        "http://localhost:8080/data/aircraft.json",
		Notifier: NotifierConfig{
			Console: true, // Default to console logging only
		},
		AlertDedupe: AlertDedupeConfig{
			Enabled:     true,
			BlockoutMin: 15 * time.Minute,
		},
		Cataloger: cataloger.ElasticSearchConfig{
			Enabled:    false, // Default to disabled
			Index:      "aircraft",
			Timeout:    30 * time.Second,
			MaxRetries: 3,
		},
	}

	// Define and parse command line flags
	flags := defineFlagsWithFlagSet(flagSet)
	// Parse the provided arguments
	if err := flagSet.Parse(args); err != nil {
		// Log error but continue with defaults
		_ = err // Suppress unused variable warning
	}
	setFlags := getSetFlagsFromFlagSet(flagSet)

	// Load from config file
	loadFromConfigFile(&cfg, flags.configPath)

	// Load from environment variables
	loadFromEnvironment(&cfg)

	// Apply command line overrides
	applyCommandLineOverrides(&cfg, flags, setFlags)

	return cfg
}

type commandLineFlags struct {
	configPath *string
	interval   *time.Duration
	radius     *float64
	altitude   *int
	lat        *float64
	lon        *float64
	dataURL    *string

	// Webhook flags
	webhookEnabled *bool
	webhookURL     *string
	webhookTimeout *time.Duration

	// RabbitMQ flags
	rabbitMQEnabled    *bool
	rabbitMQURL        *string
	rabbitMQExchange   *string
	rabbitMQRoutingKey *string
	rabbitMQTimeout    *time.Duration

	// Alert deduplication flags
	alertDedupeEnabled *bool
	alertBlockoutMin   *time.Duration

	// Cataloging flags
	catalogerEnabled    *bool
	catalogerURL        *string
	catalogerIndex      *string
	catalogerUsername   *string
	catalogerPassword   *string
	catalogerTimeout    *time.Duration
	catalogerMaxRetries *int
}

func defineFlagsWithFlagSet(flagSet *flag.FlagSet) commandLineFlags {
	flags := commandLineFlags{
		configPath: flagSet.String("config", "", "path to config file"),
		interval:   flagSet.Duration("interval", 0, "scrape interval"),
		radius:     flagSet.Float64("radius", 0, "radius of interest in km"),
		altitude:   flagSet.Int("altitude", 0, "altitude ceiling in feet"),
		lat:        flagSet.Float64("lat", 0, "base latitude"),
		lon:        flagSet.Float64("lon", 0, "base longitude"),
		dataURL:    flagSet.String("url", "", "piaware data URL"),

		// Webhook flags
		webhookEnabled: flagSet.Bool("webhook-enabled", false, "enable webhook notifications"),
		webhookURL:     flagSet.String("webhook-url", "", "webhook URL"),
		webhookTimeout: flagSet.Duration("webhook-timeout", 0, "webhook timeout"),

		// RabbitMQ flags
		rabbitMQEnabled:    flagSet.Bool("rabbitmq-enabled", false, "enable RabbitMQ notifications"),
		rabbitMQURL:        flagSet.String("rabbitmq-url", "", "RabbitMQ URL"),
		rabbitMQExchange:   flagSet.String("rabbitmq-exchange", "", "RabbitMQ exchange"),
		rabbitMQRoutingKey: flagSet.String("rabbitmq-routing-key", "", "RabbitMQ routing key"),
		rabbitMQTimeout:    flagSet.Duration("rabbitmq-timeout", 0, "RabbitMQ timeout"),

		// Alert deduplication flags
		alertDedupeEnabled: flagSet.Bool("alert-dedupe-enabled", true, "enable alert deduplication"),
		alertBlockoutMin:   flagSet.Duration("alert-blockout-min", 0, "alert blockout period"),

		// Cataloging flags
		catalogerEnabled:    flagSet.Bool("cataloger-enabled", false, "enable aircraft cataloging"),
		catalogerURL:        flagSet.String("cataloger-url", "", "ElasticSearch URL"),
		catalogerIndex:      flagSet.String("cataloger-index", "", "ElasticSearch index name"),
		catalogerUsername:   flagSet.String("cataloger-username", "", "ElasticSearch username"),
		catalogerPassword:   flagSet.String("cataloger-password", "", "ElasticSearch password"),
		catalogerTimeout:    flagSet.Duration("cataloger-timeout", 0, "ElasticSearch timeout"),
		catalogerMaxRetries: flagSet.Int("cataloger-max-retries", 0, "ElasticSearch max retries"),
	}

	return flags
}

func getSetFlagsFromFlagSet(flagSet *flag.FlagSet) map[string]bool {
	setFlags := map[string]bool{}
	flagSet.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})
	return setFlags
}

func loadFromConfigFile(cfg *Config, configPath *string) {
	path := *configPath
	if path == "" {
		path = getenv(envConfigFile, "config.json")
	}
	// #nosec G304 -- path is controlled via trusted config
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			// Ignore unmarshal errors for now
			_ = err // Suppress unused variable warning
		}
	}
}

func loadFromEnvironment(cfg *Config) {
	loadBasicConfigFromEnv(cfg)
	loadWebhookConfigFromEnv(cfg)
	loadRabbitMQConfigFromEnv(cfg)
	loadAlertDedupeConfigFromEnv(cfg)
	loadCatalogerConfigFromEnv(cfg)
}

func loadBasicConfigFromEnv(cfg *Config) {
	setDurationFromEnv(envInterval, func(d time.Duration) { cfg.ScrapeInterval = d })
	setFloatFromEnv(envRadius, func(f float64) { cfg.RadiusKm = f })
	setIntFromEnv(envAltitude, func(i int) { cfg.AltitudeMax = i })
	setFloatFromEnv(envBaseLat, func(f float64) { cfg.BaseLat = f })
	setFloatFromEnv(envBaseLon, func(f float64) { cfg.BaseLon = f })
	setStringFromEnv(envDataURL, func(s string) { cfg.DataURL = s })
}

func setDurationFromEnv(key string, setter func(time.Duration)) {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			setter(d)
		}
	}
}

func setFloatFromEnv(key string, setter func(float64)) {
	if v, ok := os.LookupEnv(key); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			setter(f)
		}
	}
}

func setIntFromEnv(key string, setter func(int)) {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			setter(i)
		}
	}
}

func setStringFromEnv(key string, setter func(string)) {
	if v, ok := os.LookupEnv(key); ok {
		setter(v)
	}
}

func loadWebhookConfigFromEnv(cfg *Config) {
	if v, ok := os.LookupEnv(envWebhookEnabled); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Notifier.Webhook.Enabled = b
		}
	}
	if v, ok := os.LookupEnv(envWebhookURL); ok {
		cfg.Notifier.Webhook.URL = v
	}
	if v, ok := os.LookupEnv(envWebhookTimeout); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Notifier.Webhook.Timeout = d
		}
	}
}

func loadRabbitMQConfigFromEnv(cfg *Config) {
	if v, ok := os.LookupEnv(envRabbitMQEnabled); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Notifier.RabbitMQ.Enabled = b
		}
	}
	if v, ok := os.LookupEnv(envRabbitMQURL); ok {
		cfg.Notifier.RabbitMQ.URL = v
	}
	if v, ok := os.LookupEnv(envRabbitMQExchange); ok {
		cfg.Notifier.RabbitMQ.Exchange = v
	}
	if v, ok := os.LookupEnv(envRabbitMQRoutingKey); ok {
		cfg.Notifier.RabbitMQ.RoutingKey = v
	}
	if v, ok := os.LookupEnv(envRabbitMQTimeout); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Notifier.RabbitMQ.Timeout = d
		}
	}
}

func loadAlertDedupeConfigFromEnv(cfg *Config) {
	if v, ok := os.LookupEnv(envAlertDedupeEnabled); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.AlertDedupe.Enabled = b
		}
	}
	if v, ok := os.LookupEnv(envAlertBlockoutMin); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.AlertDedupe.BlockoutMin = d
		}
	}
}

func loadCatalogerConfigFromEnv(cfg *Config) {
	if v, ok := os.LookupEnv(envCatalogerEnabled); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Cataloger.Enabled = b
		}
	}
	if v, ok := os.LookupEnv(envCatalogerURL); ok {
		cfg.Cataloger.URL = v
	}
	if v, ok := os.LookupEnv(envCatalogerIndex); ok {
		cfg.Cataloger.Index = v
	}
	if v, ok := os.LookupEnv(envCatalogerUsername); ok {
		cfg.Cataloger.Username = v
	}
	if v, ok := os.LookupEnv(envCatalogerPassword); ok {
		cfg.Cataloger.Password = v
	}
	if v, ok := os.LookupEnv(envCatalogerTimeout); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Cataloger.Timeout = d
		}
	}
	if v, ok := os.LookupEnv(envCatalogerMaxRetries); ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Cataloger.MaxRetries = i
		}
	}
}

func applyCommandLineOverrides(cfg *Config, flags commandLineFlags, setFlags map[string]bool) {
	applyBasicCommandLineOverrides(cfg, flags, setFlags)
	applyWebhookCommandLineOverrides(cfg, flags, setFlags)
	applyRabbitMQCommandLineOverrides(cfg, flags, setFlags)
	applyAlertDedupeCommandLineOverrides(cfg, flags, setFlags)
	applyCatalogerCommandLineOverrides(cfg, flags, setFlags)
}

func applyBasicCommandLineOverrides(cfg *Config, flags commandLineFlags, setFlags map[string]bool) {
	if setFlags["interval"] {
		cfg.ScrapeInterval = *flags.interval
	}
	if setFlags["radius"] {
		cfg.RadiusKm = *flags.radius
	}
	if setFlags["altitude"] {
		cfg.AltitudeMax = *flags.altitude
	}
	if setFlags["lat"] {
		cfg.BaseLat = *flags.lat
	}
	if setFlags["lon"] {
		cfg.BaseLon = *flags.lon
	}
	if setFlags["url"] {
		cfg.DataURL = *flags.dataURL
	}
}

func applyWebhookCommandLineOverrides(cfg *Config, flags commandLineFlags, setFlags map[string]bool) {
	if setFlags["webhook-enabled"] {
		cfg.Notifier.Webhook.Enabled = *flags.webhookEnabled
	}
	if setFlags["webhook-url"] {
		cfg.Notifier.Webhook.URL = *flags.webhookURL
	}
	if setFlags["webhook-timeout"] {
		cfg.Notifier.Webhook.Timeout = *flags.webhookTimeout
	}
}

func applyRabbitMQCommandLineOverrides(cfg *Config, flags commandLineFlags, setFlags map[string]bool) {
	if setFlags["rabbitmq-enabled"] {
		cfg.Notifier.RabbitMQ.Enabled = *flags.rabbitMQEnabled
	}
	if setFlags["rabbitmq-url"] {
		cfg.Notifier.RabbitMQ.URL = *flags.rabbitMQURL
	}
	if setFlags["rabbitmq-exchange"] {
		cfg.Notifier.RabbitMQ.Exchange = *flags.rabbitMQExchange
	}
	if setFlags["rabbitmq-routing-key"] {
		cfg.Notifier.RabbitMQ.RoutingKey = *flags.rabbitMQRoutingKey
	}
	if setFlags["rabbitmq-timeout"] {
		cfg.Notifier.RabbitMQ.Timeout = *flags.rabbitMQTimeout
	}
}

func applyAlertDedupeCommandLineOverrides(cfg *Config, flags commandLineFlags, setFlags map[string]bool) {
	if setFlags["alert-dedupe-enabled"] {
		cfg.AlertDedupe.Enabled = *flags.alertDedupeEnabled
	}
	if setFlags["alert-blockout-min"] {
		cfg.AlertDedupe.BlockoutMin = *flags.alertBlockoutMin
	}
}

func applyCatalogerCommandLineOverrides(cfg *Config, flags commandLineFlags, setFlags map[string]bool) {
	if setFlags["cataloger-enabled"] {
		cfg.Cataloger.Enabled = *flags.catalogerEnabled
	}
	if setFlags["cataloger-url"] {
		cfg.Cataloger.URL = *flags.catalogerURL
	}
	if setFlags["cataloger-index"] {
		cfg.Cataloger.Index = *flags.catalogerIndex
	}
	if setFlags["cataloger-username"] {
		cfg.Cataloger.Username = *flags.catalogerUsername
	}
	if setFlags["cataloger-password"] {
		cfg.Cataloger.Password = *flags.catalogerPassword
	}
	if setFlags["cataloger-timeout"] {
		cfg.Cataloger.Timeout = *flags.catalogerTimeout
	}
	if setFlags["cataloger-max-retries"] {
		cfg.Cataloger.MaxRetries = *flags.catalogerMaxRetries
	}
}

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
