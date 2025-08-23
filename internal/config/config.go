package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"time"
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
)

// Load reads configuration from config file, environment variables and command line flags
// in order of increasing precedence.
func Load() Config {
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
	}

	// Command line flags
	configPath := flag.String("config", "", "path to config file")

	interval := flag.Duration("interval", 0, "scrape interval")
	radius := flag.Float64("radius", 0, "radius of interest in km")
	altitude := flag.Int("altitude", 0, "altitude ceiling in feet")
	lat := flag.Float64("lat", 0, "base latitude")
	lon := flag.Float64("lon", 0, "base longitude")
	dataURL := flag.String("url", "", "piaware data URL")

	// Webhook flags
	webhookEnabled := flag.Bool("webhook-enabled", false, "enable webhook notifications")
	webhookURL := flag.String("webhook-url", "", "webhook URL")
	webhookTimeout := flag.Duration("webhook-timeout", 0, "webhook timeout")

	// RabbitMQ flags
	rabbitMQEnabled := flag.Bool("rabbitmq-enabled", false, "enable RabbitMQ notifications")
	rabbitMQURL := flag.String("rabbitmq-url", "", "RabbitMQ URL")
	rabbitMQExchange := flag.String("rabbitmq-exchange", "", "RabbitMQ exchange")
	rabbitMQRoutingKey := flag.String("rabbitmq-routing-key", "", "RabbitMQ routing key")
	rabbitMQTimeout := flag.Duration("rabbitmq-timeout", 0, "RabbitMQ timeout")

	// Alert deduplication flags
	alertDedupeEnabled := flag.Bool("alert-dedupe-enabled", true, "enable alert deduplication")
	alertBlockoutMin := flag.Duration("alert-blockout-min", 0, "alert blockout period")

	flag.Parse()

	setFlags := map[string]bool{}
	flag.CommandLine.Visit(func(f *flag.Flag) { setFlags[f.Name] = true })

	// Config file
	path := *configPath
	if path == "" {
		path = getenv(envConfigFile, "config.json")
	}
	// #nosec G304 -- path is controlled via trusted config
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	// Environment variables
	if v, ok := os.LookupEnv(envInterval); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.ScrapeInterval = d
		}
	}
	if v, ok := os.LookupEnv(envRadius); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.RadiusKm = f
		}
	}
	if v, ok := os.LookupEnv(envAltitude); ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.AltitudeMax = i
		}
	}
	if v, ok := os.LookupEnv(envBaseLat); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.BaseLat = f
		}
	}
	if v, ok := os.LookupEnv(envBaseLon); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.BaseLon = f
		}
	}
	if v, ok := os.LookupEnv(envDataURL); ok {
		cfg.DataURL = v
	}

	// Webhook environment variables
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

	// RabbitMQ environment variables
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

	// Alert deduplication environment variables
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

	// Command line overrides
	if setFlags["interval"] {
		cfg.ScrapeInterval = *interval
	}
	if setFlags["radius"] {
		cfg.RadiusKm = *radius
	}
	if setFlags["altitude"] {
		cfg.AltitudeMax = *altitude
	}
	if setFlags["lat"] {
		cfg.BaseLat = *lat
	}
	if setFlags["lon"] {
		cfg.BaseLon = *lon
	}
	if setFlags["url"] {
		cfg.DataURL = *dataURL
	}

	// Webhook command line overrides
	if setFlags["webhook-enabled"] {
		cfg.Notifier.Webhook.Enabled = *webhookEnabled
	}
	if setFlags["webhook-url"] {
		cfg.Notifier.Webhook.URL = *webhookURL
	}
	if setFlags["webhook-timeout"] {
		cfg.Notifier.Webhook.Timeout = *webhookTimeout
	}

	// RabbitMQ command line overrides
	if setFlags["rabbitmq-enabled"] {
		cfg.Notifier.RabbitMQ.Enabled = *rabbitMQEnabled
	}
	if setFlags["rabbitmq-url"] {
		cfg.Notifier.RabbitMQ.URL = *rabbitMQURL
	}
	if setFlags["rabbitmq-exchange"] {
		cfg.Notifier.RabbitMQ.Exchange = *rabbitMQExchange
	}
	if setFlags["rabbitmq-routing-key"] {
		cfg.Notifier.RabbitMQ.RoutingKey = *rabbitMQRoutingKey
	}
	if setFlags["rabbitmq-timeout"] {
		cfg.Notifier.RabbitMQ.Timeout = *rabbitMQTimeout
	}

	// Alert deduplication command line overrides
	if setFlags["alert-dedupe-enabled"] {
		cfg.AlertDedupe.Enabled = *alertDedupeEnabled
	}
	if setFlags["alert-blockout-min"] {
		cfg.AlertDedupe.BlockoutMin = *alertBlockoutMin
	}

	return cfg
}

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
