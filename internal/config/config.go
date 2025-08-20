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
}

// NotifierConfig holds notifier settings.
type NotifierConfig struct {
	Method string
	Email  EmailConfig
}

// EmailConfig holds email notifier settings.
type EmailConfig struct {
	SMTPServer string
	SMTPPort   int
	Username   string
	Password   string
	From       string
	To         string
}

const (
	envConfigFile = "WFO_CONFIG"
	envInterval   = "WFO_INTERVAL"
	envRadius     = "WFO_RADIUS_KM"
	envAltitude   = "WFO_ALTITUDE_MAX"
	envBaseLat    = "WFO_BASE_LAT"
	envBaseLon    = "WFO_BASE_LON"
	envDataURL    = "WFO_DATA_URL"
	envNotify     = "WFO_NOTIFY"
	envSMTPServer = "WFO_SMTP_SERVER"
	envSMTPPort   = "WFO_SMTP_PORT"
	envSMTPUser   = "WFO_SMTP_USER"
	envSMTPPass   = "WFO_SMTP_PASS" // #nosec G101
	envEmailFrom  = "WFO_EMAIL_FROM"
	envEmailTo    = "WFO_EMAIL_TO"
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
			Method: "email",
			Email:  EmailConfig{SMTPPort: 587},
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
	notify := flag.String("notify", "", "notification method")

	smtpServer := flag.String("smtp-server", "", "SMTP server")
	smtpPort := flag.Int("smtp-port", 0, "SMTP port")
	smtpUser := flag.String("smtp-user", "", "SMTP username")
	smtpPass := flag.String("smtp-pass", "", "SMTP password")
	emailFrom := flag.String("email-from", "", "email sender")
	emailTo := flag.String("email-to", "", "email recipient")

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
	if v, ok := os.LookupEnv(envNotify); ok {
		cfg.Notifier.Method = v
	}
	if v, ok := os.LookupEnv(envSMTPServer); ok {
		cfg.Notifier.Email.SMTPServer = v
	}
	if v, ok := os.LookupEnv(envSMTPPort); ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Notifier.Email.SMTPPort = i
		}
	}
	if v, ok := os.LookupEnv(envSMTPUser); ok {
		cfg.Notifier.Email.Username = v
	}
	if v, ok := os.LookupEnv(envSMTPPass); ok {
		cfg.Notifier.Email.Password = v
	}
	if v, ok := os.LookupEnv(envEmailFrom); ok {
		cfg.Notifier.Email.From = v
	}
	if v, ok := os.LookupEnv(envEmailTo); ok {
		cfg.Notifier.Email.To = v
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
	if setFlags["notify"] {
		cfg.Notifier.Method = *notify
	}
	if setFlags["smtp-server"] {
		cfg.Notifier.Email.SMTPServer = *smtpServer
	}
	if setFlags["smtp-port"] {
		cfg.Notifier.Email.SMTPPort = *smtpPort
	}
	if setFlags["smtp-user"] {
		cfg.Notifier.Email.Username = *smtpUser
	}
	if setFlags["smtp-pass"] {
		cfg.Notifier.Email.Password = *smtpPass
	}
	if setFlags["email-from"] {
		cfg.Notifier.Email.From = *emailFrom
	}
	if setFlags["email-to"] {
		cfg.Notifier.Email.To = *emailTo
	}

	return cfg
}

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
