package config

import (
	"flag"
	"os"
	"testing"
	"time"
)

// helper to reset environment and flags
func reset() {
	os.Clearenv()
	// Reset the global flag set
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = []string{os.Args[0]}
}

func TestLoadPrecedence(t *testing.T) {
	reset()
	// create temp config file
	cfgFile, err := os.CreateTemp(t.TempDir(), "cfg*.json")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	if _, err := cfgFile.WriteString(`{"RadiusKm":10}`); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := cfgFile.Close(); err != nil {
		t.Fatalf("close config: %v", err)
	}

	if err := os.Setenv("WFO_CONFIG", cfgFile.Name()); err != nil {
		t.Fatalf("set env: %v", err)
	}
	if err := os.Setenv("WFO_RADIUS_KM", "20"); err != nil {
		t.Fatalf("set env: %v", err)
	}

	// Create a test flag set
	testFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)

	// set command line flag - command line should override env and config file
	args := []string{"-radius", "30"}

	cfg := LoadWithFlagSetAndArgs(testFlagSet, args)
	if cfg.RadiusKm != 30 {
		t.Fatalf("expected radius 30 (command line override), got %v", cfg.RadiusKm)
	}
}

func TestLoadDefaults(t *testing.T) {
	reset()
	cfg := Load()
	if cfg.ScrapeInterval != time.Minute {
		t.Errorf("expected default interval %v, got %v", time.Minute, cfg.ScrapeInterval)
	}
	if cfg.RadiusKm != 25 {
		t.Errorf("expected default radius 25, got %v", cfg.RadiusKm)
	}
	if cfg.AltitudeMax != 10000 {
		t.Errorf("expected default altitude 10000, got %v", cfg.AltitudeMax)
	}
	if cfg.DataURL != "http://localhost:8080/data/aircraft.json" {
		t.Errorf("expected default data URL, got %v", cfg.DataURL)
	}
	if !cfg.Notifier.Console {
		t.Error("expected console notifier to be enabled by default")
	}
	if !cfg.AlertDedupe.Enabled {
		t.Error("expected alert deduplication to be enabled by default")
	}
	if cfg.AlertDedupe.BlockoutMin != 15*time.Minute {
		t.Errorf("expected default blockout period %v, got %v", 15*time.Minute, cfg.AlertDedupe.BlockoutMin)
	}
	if cfg.Cataloger.Enabled {
		t.Error("expected cataloger to be disabled by default")
	}
	if cfg.Cataloger.Index != "aircraft" {
		t.Errorf("expected default index 'aircraft', got %v", cfg.Cataloger.Index)
	}
}

func TestLoadFromConfigFile(t *testing.T) {
	reset()
	cfgFile, err := os.CreateTemp(t.TempDir(), "cfg*.json")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer func() {
		if removeErr := os.Remove(cfgFile.Name()); removeErr != nil {
			t.Logf("Failed to remove temp file: %v", removeErr)
		}
	}()

	configData := `{
		"ScrapeInterval": "2m",
		"RadiusKm": 50.0,
		"AltitudeMax": 15000,
		"BaseLat": 40.7128,
		"BaseLon": -74.0060,
		"DataURL": "http://example.com/data.json",
		"Notifier": {
			"Webhook": {
				"Enabled": true,
				"URL": "http://webhook.example.com",
				"Timeout": "10s"
			},
			"RabbitMQ": {
				"Enabled": true,
				"URL": "amqp://localhost:5672",
				"Exchange": "aircraft",
				"RoutingKey": "alerts",
				"Timeout": "5s"
			},
			"Console": false
		},
		"AlertDedupe": {
			"Enabled": false,
			"BlockoutMin": "30m"
		},
		"Cataloger": {
			"Enabled": true,
			"URL": "http://elasticsearch:9200",
			"Index": "test-aircraft",
			"Username": "testuser",
			"Password": "testpass",
			"Timeout": "60s",
			"MaxRetries": 5
		}
	}`

	if _, err := cfgFile.WriteString(configData); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := cfgFile.Close(); err != nil {
		t.Fatalf("close config: %v", err)
	}

	if err := os.Setenv("WFO_CONFIG", cfgFile.Name()); err != nil {
		t.Fatalf("set env: %v", err)
	}

	cfg := Load()

	// Verify config file values were loaded
	if cfg.ScrapeInterval != 2*time.Minute {
		t.Errorf("expected interval 2m, got %v", cfg.ScrapeInterval)
	}
	if cfg.RadiusKm != 50.0 {
		t.Errorf("expected radius 50.0, got %v", cfg.RadiusKm)
	}
	if cfg.AltitudeMax != 15000 {
		t.Errorf("expected altitude 15000, got %v", cfg.AltitudeMax)
	}
	if cfg.BaseLat != 40.7128 {
		t.Errorf("expected lat 40.7128, got %v", cfg.BaseLat)
	}
	if cfg.BaseLon != -74.0060 {
		t.Errorf("expected lon -74.0060, got %v", cfg.BaseLon)
	}
	if cfg.DataURL != "http://example.com/data.json" {
		t.Errorf("expected data URL, got %v", cfg.DataURL)
	}
	if !cfg.Notifier.Webhook.Enabled {
		t.Error("expected webhook to be enabled")
	}
	if cfg.Notifier.Webhook.URL != "http://webhook.example.com" {
		t.Errorf("expected webhook URL, got %v", cfg.Notifier.Webhook.URL)
	}
	if cfg.Notifier.Webhook.Timeout != 10*time.Second {
		t.Errorf("expected webhook timeout 10s, got %v", cfg.Notifier.Webhook.Timeout)
	}
	if !cfg.Notifier.RabbitMQ.Enabled {
		t.Error("expected RabbitMQ to be enabled")
	}
	if cfg.Notifier.RabbitMQ.URL != "amqp://localhost:5672" {
		t.Errorf("expected RabbitMQ URL, got %v", cfg.Notifier.RabbitMQ.URL)
	}
	if cfg.Notifier.RabbitMQ.Exchange != "aircraft" {
		t.Errorf("expected RabbitMQ exchange, got %v", cfg.Notifier.RabbitMQ.Exchange)
	}
	if cfg.Notifier.RabbitMQ.RoutingKey != "alerts" {
		t.Errorf("expected RabbitMQ routing key, got %v", cfg.Notifier.RabbitMQ.RoutingKey)
	}
	if cfg.Notifier.RabbitMQ.Timeout != 5*time.Second {
		t.Errorf("expected RabbitMQ timeout 5s, got %v", cfg.Notifier.RabbitMQ.Timeout)
	}
	if cfg.Notifier.Console {
		t.Error("expected console to be disabled")
	}
	if cfg.AlertDedupe.Enabled {
		t.Error("expected alert deduplication to be disabled")
	}
	if cfg.AlertDedupe.BlockoutMin != 30*time.Minute {
		t.Errorf("expected blockout period 30m, got %v", cfg.AlertDedupe.BlockoutMin)
	}
	if !cfg.Cataloger.Enabled {
		t.Error("expected cataloger to be enabled")
	}
	if cfg.Cataloger.URL != "http://elasticsearch:9200" {
		t.Errorf("expected cataloger URL, got %v", cfg.Cataloger.URL)
	}
	if cfg.Cataloger.Index != "test-aircraft" {
		t.Errorf("expected cataloger index, got %v", cfg.Cataloger.Index)
	}
	if cfg.Cataloger.Username != "testuser" {
		t.Errorf("expected cataloger username, got %v", cfg.Cataloger.Username)
	}
	if cfg.Cataloger.Password != "testpass" {
		t.Errorf("expected cataloger password, got %v", cfg.Cataloger.Password)
	}
	if cfg.Cataloger.Timeout != 60*time.Second {
		t.Errorf("expected cataloger timeout 60s, got %v", cfg.Cataloger.Timeout)
	}
	if cfg.Cataloger.MaxRetries != 5 {
		t.Errorf("expected cataloger max retries 5, got %v", cfg.Cataloger.MaxRetries)
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	reset()

	// Set environment variables
	envVars := map[string]string{
		"WFO_INTERVAL":              "5m",
		"WFO_RADIUS_KM":             "100.5",
		"WFO_ALTITUDE_MAX":          "20000",
		"WFO_BASE_LAT":              "51.5074",
		"WFO_BASE_LON":              "-0.1278",
		"WFO_DATA_URL":              "http://env.example.com/data.json",
		"WFO_WEBHOOK_ENABLED":       "true",
		"WFO_WEBHOOK_URL":           "http://env.webhook.example.com",
		"WFO_WEBHOOK_TIMEOUT":       "15s",
		"WFO_RABBITMQ_ENABLED":      "true",
		"WFO_RABBITMQ_URL":          "amqp://env.rabbitmq:5672",
		"WFO_RABBITMQ_EXCHANGE":     "env-aircraft",
		"WFO_RABBITMQ_ROUTING_KEY":  "env-alerts",
		"WFO_RABBITMQ_TIMEOUT":      "8s",
		"WFO_ALERT_DEDUPE_ENABLED":  "false",
		"WFO_ALERT_BLOCKOUT_MIN":    "45m",
		"WFO_CATALOGER_ENABLED":     "true",
		"WFO_CATALOGER_URL":         "http://env.elasticsearch:9200",
		"WFO_CATALOGER_INDEX":       "env-aircraft",
		"WFO_CATALOGER_USERNAME":    "envuser",
		"WFO_CATALOGER_PASSWORD":    "envpass",
		"WFO_CATALOGER_TIMEOUT":     "90s",
		"WFO_CATALOGER_MAX_RETRIES": "7",
	}

	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("set env %s: %v", key, err)
		}
	}

	cfg := Load()

	// Verify environment variables were loaded
	if cfg.ScrapeInterval != 5*time.Minute {
		t.Errorf("expected interval 5m, got %v", cfg.ScrapeInterval)
	}
	if cfg.RadiusKm != 100.5 {
		t.Errorf("expected radius 100.5, got %v", cfg.RadiusKm)
	}
	if cfg.AltitudeMax != 20000 {
		t.Errorf("expected altitude 20000, got %v", cfg.AltitudeMax)
	}
	if cfg.BaseLat != 51.5074 {
		t.Errorf("expected lat 51.5074, got %v", cfg.BaseLat)
	}
	if cfg.BaseLon != -0.1278 {
		t.Errorf("expected lon -0.1278, got %v", cfg.BaseLon)
	}
	if cfg.DataURL != "http://env.example.com/data.json" {
		t.Errorf("expected data URL, got %v", cfg.DataURL)
	}
	if !cfg.Notifier.Webhook.Enabled {
		t.Error("expected webhook to be enabled")
	}
	if cfg.Notifier.Webhook.URL != "http://env.webhook.example.com" {
		t.Errorf("expected webhook URL, got %v", cfg.Notifier.Webhook.URL)
	}
	if cfg.Notifier.Webhook.Timeout != 15*time.Second {
		t.Errorf("expected webhook timeout 15s, got %v", cfg.Notifier.Webhook.Timeout)
	}
	if !cfg.Notifier.RabbitMQ.Enabled {
		t.Error("expected RabbitMQ to be enabled")
	}
	if cfg.Notifier.RabbitMQ.URL != "amqp://env.rabbitmq:5672" {
		t.Errorf("expected RabbitMQ URL, got %v", cfg.Notifier.RabbitMQ.URL)
	}
	if cfg.Notifier.RabbitMQ.Exchange != "env-aircraft" {
		t.Errorf("expected RabbitMQ exchange, got %v", cfg.Notifier.RabbitMQ.Exchange)
	}
	if cfg.Notifier.RabbitMQ.RoutingKey != "env-alerts" {
		t.Errorf("expected RabbitMQ routing key, got %v", cfg.Notifier.RabbitMQ.RoutingKey)
	}
	if cfg.Notifier.RabbitMQ.Timeout != 8*time.Second {
		t.Errorf("expected RabbitMQ timeout 8s, got %v", cfg.Notifier.RabbitMQ.Timeout)
	}
	if cfg.AlertDedupe.Enabled {
		t.Error("expected alert deduplication to be disabled")
	}
	if cfg.AlertDedupe.BlockoutMin != 45*time.Minute {
		t.Errorf("expected blockout period 45m, got %v", cfg.AlertDedupe.BlockoutMin)
	}
	if !cfg.Cataloger.Enabled {
		t.Error("expected cataloger to be enabled")
	}
	if cfg.Cataloger.URL != "http://env.elasticsearch:9200" {
		t.Errorf("expected cataloger URL, got %v", cfg.Cataloger.URL)
	}
	if cfg.Cataloger.Index != "env-aircraft" {
		t.Errorf("expected cataloger index, got %v", cfg.Cataloger.Index)
	}
	if cfg.Cataloger.Username != "envuser" {
		t.Errorf("expected cataloger username, got %v", cfg.Cataloger.Username)
	}
	if cfg.Cataloger.Password != "envpass" {
		t.Errorf("expected cataloger password, got %v", cfg.Cataloger.Password)
	}
	if cfg.Cataloger.Timeout != 90*time.Second {
		t.Errorf("expected cataloger timeout 90s, got %v", cfg.Cataloger.Timeout)
	}
	if cfg.Cataloger.MaxRetries != 7 {
		t.Errorf("expected cataloger max retries 7, got %v", cfg.Cataloger.MaxRetries)
	}
}

func TestLoadFromCommandLine(t *testing.T) {
	reset()

	// Create a test flag set
	testFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)

	// Set command line flags - use correct Go flag syntax for boolean flags
	args := []string{
		"-interval", "10m",
		"-radius", "75.25",
		"-altitude", "25000",
		"-lat", "35.6762",
		"-lon", "139.6503",
		"-url", "http://cli.example.com/data.json",
		"-webhook-enabled", // boolean flag - no value needed
		"-webhook-url", "http://cli.webhook.example.com",
		"-webhook-timeout", "20s",
		"-rabbitmq-enabled", // boolean flag - no value needed
		"-rabbitmq-url", "amqp://cli.rabbitmq:5672",
		"-rabbitmq-exchange", "cli-aircraft",
		"-rabbitmq-routing-key", "cli-alerts",
		"-rabbitmq-timeout", "12s",
		"-alert-dedupe-enabled=false", // boolean flag with explicit value
		"-alert-blockout-min", "60m",
		"-cataloger-enabled", // boolean flag - no value needed
		"-cataloger-url", "http://cli.elasticsearch:9200",
		"-cataloger-index", "cli-aircraft",
		"-cataloger-username", "cliuser",
		"-cataloger-password", "clipass",
		"-cataloger-timeout", "120s",
		"-cataloger-max-retries", "10",
	}

	cfg := LoadWithFlagSetAndArgs(testFlagSet, args)

	// Verify command line flags were loaded
	if cfg.ScrapeInterval != 10*time.Minute {
		t.Errorf("expected interval 10m, got %v", cfg.ScrapeInterval)
	}
	if cfg.RadiusKm != 75.25 {
		t.Errorf("expected radius 75.25, got %v", cfg.RadiusKm)
	}
	if cfg.AltitudeMax != 25000 {
		t.Errorf("expected altitude 25000, got %v", cfg.AltitudeMax)
	}
	if cfg.BaseLat != 35.6762 {
		t.Errorf("expected lat 35.6762, got %v", cfg.BaseLat)
	}
	if cfg.BaseLon != 139.6503 {
		t.Errorf("expected lon 139.6503, got %v", cfg.BaseLon)
	}
	if cfg.DataURL != "http://cli.example.com/data.json" {
		t.Errorf("expected data URL, got %v", cfg.DataURL)
	}
	if !cfg.Notifier.Webhook.Enabled {
		t.Error("expected webhook to be enabled")
	}
	if cfg.Notifier.Webhook.URL != "http://cli.webhook.example.com" {
		t.Errorf("expected webhook URL, got %v", cfg.Notifier.Webhook.URL)
	}
	if cfg.Notifier.Webhook.Timeout != 20*time.Second {
		t.Errorf("expected webhook timeout 20s, got %v", cfg.Notifier.Webhook.Timeout)
	}
	if !cfg.Notifier.RabbitMQ.Enabled {
		t.Error("expected RabbitMQ to be enabled")
	}
	if cfg.Notifier.RabbitMQ.URL != "amqp://cli.rabbitmq:5672" {
		t.Errorf("expected RabbitMQ URL, got %v", cfg.Notifier.RabbitMQ.URL)
	}
	if cfg.Notifier.RabbitMQ.Exchange != "cli-aircraft" {
		t.Errorf("expected RabbitMQ exchange, got %v", cfg.Notifier.RabbitMQ.Exchange)
	}
	if cfg.Notifier.RabbitMQ.RoutingKey != "cli-alerts" {
		t.Errorf("expected RabbitMQ routing key, got %v", cfg.Notifier.RabbitMQ.RoutingKey)
	}
	if cfg.Notifier.RabbitMQ.Timeout != 12*time.Second {
		t.Errorf("expected RabbitMQ timeout 12s, got %v", cfg.Notifier.RabbitMQ.Timeout)
	}
	if cfg.AlertDedupe.Enabled {
		t.Error("expected alert deduplication to be disabled")
	}
	if cfg.AlertDedupe.BlockoutMin != 60*time.Minute {
		t.Errorf("expected blockout period 60m, got %v", cfg.AlertDedupe.BlockoutMin)
	}
	if !cfg.Cataloger.Enabled {
		t.Error("expected cataloger to be enabled")
	}
	if cfg.Cataloger.URL != "http://cli.elasticsearch:9200" {
		t.Errorf("expected cataloger URL, got %v", cfg.Cataloger.URL)
	}
	if cfg.Cataloger.Index != "cli-aircraft" {
		t.Errorf("expected cataloger index, got %v", cfg.Cataloger.Index)
	}
	if cfg.Cataloger.Username != "cliuser" {
		t.Errorf("expected cataloger username, got %v", cfg.Cataloger.Username)
	}
	if cfg.Cataloger.Password != "clipass" {
		t.Errorf("expected cataloger password, got %v", cfg.Cataloger.Password)
	}
	if cfg.Cataloger.Timeout != 120*time.Second {
		t.Errorf("expected cataloger timeout 120s, got %v", cfg.Cataloger.Timeout)
	}
	if cfg.Cataloger.MaxRetries != 10 {
		t.Errorf("expected cataloger max retries 10, got %v", cfg.Cataloger.MaxRetries)
	}
}

func TestHelperFunctions(t *testing.T) {
	reset()

	// Test setDurationFromEnv
	t.Run("setDurationFromEnv", func(t *testing.T) {
		var result time.Duration
		setDurationFromEnv("TEST_DURATION", func(d time.Duration) { result = d })
		if result != 0 {
			t.Errorf("expected 0 duration when env not set, got %v", result)
		}

		if err := os.Setenv("TEST_DURATION", "25s"); err != nil {
			t.Fatalf("set env: %v", err)
		}
		setDurationFromEnv("TEST_DURATION", func(d time.Duration) { result = d })
		if result != 25*time.Second {
			t.Errorf("expected 25s duration, got %v", result)
		}

		// Test invalid duration
		if err := os.Setenv("TEST_DURATION", "invalid"); err != nil {
			t.Fatalf("set env: %v", err)
		}
		setDurationFromEnv("TEST_DURATION", func(d time.Duration) { result = d })
		if result != 25*time.Second { // Should not change with invalid input
			t.Errorf("expected duration to remain unchanged, got %v", result)
		}
	})

	// Test setFloatFromEnv
	t.Run("setFloatFromEnv", func(t *testing.T) {
		var result float64
		setFloatFromEnv("TEST_FLOAT", func(f float64) { result = f })
		if result != 0 {
			t.Errorf("expected 0 float when env not set, got %v", result)
		}

		if err := os.Setenv("TEST_FLOAT", "42.5"); err != nil {
			t.Fatalf("set env: %v", err)
		}
		setFloatFromEnv("TEST_FLOAT", func(f float64) { result = f })
		if result != 42.5 {
			t.Errorf("expected 42.5 float, got %v", result)
		}

		// Test invalid float
		if err := os.Setenv("TEST_FLOAT", "invalid"); err != nil {
			t.Fatalf("set env: %v", err)
		}
		setFloatFromEnv("TEST_FLOAT", func(f float64) { result = f })
		if result != 42.5 { // Should not change with invalid input
			t.Errorf("expected float to remain unchanged, got %v", result)
		}
	})

	// Test setIntFromEnv
	t.Run("setIntFromEnv", func(t *testing.T) {
		var result int
		setIntFromEnv("TEST_INT", func(i int) { result = i })
		if result != 0 {
			t.Errorf("expected 0 int when env not set, got %v", result)
		}

		if err := os.Setenv("TEST_INT", "123"); err != nil {
			t.Fatalf("set env: %v", err)
		}
		setIntFromEnv("TEST_INT", func(i int) { result = i })
		if result != 123 {
			t.Errorf("expected 123 int, got %v", result)
		}

		// Test invalid int
		if err := os.Setenv("TEST_INT", "invalid"); err != nil {
			t.Fatalf("set env: %v", err)
		}
		setIntFromEnv("TEST_INT", func(i int) { result = i })
		if result != 123 { // Should not change with invalid input
			t.Errorf("expected int to remain unchanged, got %v", result)
		}
	})

	// Test setStringFromEnv
	t.Run("setStringFromEnv", func(t *testing.T) {
		var result string
		setStringFromEnv("TEST_STRING", func(s string) { result = s })
		if result != "" {
			t.Errorf("expected empty string when env not set, got %v", result)
		}

		if err := os.Setenv("TEST_STRING", "test-value"); err != nil {
			t.Fatalf("set env: %v", err)
		}
		setStringFromEnv("TEST_STRING", func(s string) { result = s })
		if result != "test-value" {
			t.Errorf("expected 'test-value' string, got %v", result)
		}
	})
}

func TestConfigFileNotFound(t *testing.T) {
	reset()
	if err := os.Setenv("WFO_CONFIG", "/nonexistent/file.json"); err != nil {
		t.Fatalf("set env: %v", err)
	}

	// Should not panic and should use defaults
	cfg := Load()
	if cfg.ScrapeInterval != time.Minute {
		t.Errorf("expected default interval when config file not found, got %v", cfg.ScrapeInterval)
	}
}

func TestInvalidEnvironmentValues(t *testing.T) {
	reset()

	// Set invalid environment variables
	invalidEnvVars := map[string]string{
		"WFO_INTERVAL":     "invalid-duration",
		"WFO_RADIUS_KM":    "not-a-number",
		"WFO_ALTITUDE_MAX": "not-an-integer",
	}

	for key, value := range invalidEnvVars {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("set env %s: %v", key, value)
		}
	}

	cfg := Load()

	// Should use defaults when invalid values are provided
	if cfg.ScrapeInterval != time.Minute {
		t.Errorf("expected default interval with invalid env, got %v", cfg.ScrapeInterval)
	}
	if cfg.RadiusKm != 25.0 {
		t.Errorf("expected default radius with invalid env, got %v", cfg.RadiusKm)
	}
	if cfg.AltitudeMax != 10000 {
		t.Errorf("expected default altitude with invalid env, got %v", cfg.AltitudeMax)
	}
}

func TestFlagParsing(t *testing.T) {
	reset()

	// Create a test flag set
	testFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)

	// Define flags
	flags := defineFlagsWithFlagSet(testFlagSet)

	// Test with non-boolean flags first
	t.Run("NonBooleanFlags", func(t *testing.T) {
		// Reset the flag set
		testFlagSet = flag.NewFlagSet("test", flag.ContinueOnError)
		flags = defineFlagsWithFlagSet(testFlagSet)

		// Parse non-boolean flags
		args := []string{"-webhook-url", "http://test.com", "-webhook-timeout", "30s"}
		t.Logf("Parsing args: %v", args)

		err := testFlagSet.Parse(args)
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}

		// Debug: check what the flag set thinks it parsed
		t.Logf("Flag set args: %v", testFlagSet.Args())
		t.Logf("Flag set NFlag: %d", testFlagSet.NFlag())

		// Check which flags were set
		setFlags := getSetFlagsFromFlagSet(testFlagSet)

		// Debug: print all set flags
		for name, set := range setFlags {
			t.Logf("Flag '%s' is set: %v", name, set)
		}

		// Debug: print the actual flag values
		t.Logf("webhook-url flag value: %v", *flags.webhookURL)
		t.Logf("webhook-timeout flag value: %v", *flags.webhookTimeout)

		// Verify flags were parsed
		if !setFlags["webhook-url"] {
			t.Error("webhook-url flag should be set")
		}
		if !setFlags["webhook-timeout"] {
			t.Error("webhook-timeout flag should be set")
		}

		// Verify the flag values
		if *flags.webhookURL != "http://test.com" {
			t.Errorf("Expected webhook-url to be 'http://test.com', got %v", *flags.webhookURL)
		}
		if *flags.webhookTimeout != 30*time.Second {
			t.Errorf("Expected webhook-timeout to be 30s, got %v", *flags.webhookTimeout)
		}
	})

	// Test with boolean flags
	t.Run("BooleanFlags", func(t *testing.T) {
		// Reset the flag set
		testFlagSet = flag.NewFlagSet("test", flag.ContinueOnError)
		flags = defineFlagsWithFlagSet(testFlagSet)

		// Parse boolean flags - use correct Go flag syntax
		args := []string{"-webhook-enabled", "-webhook-url", "http://test.com"}
		t.Logf("Parsing args: %v", args)

		err := testFlagSet.Parse(args)
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}

		// Debug: check what the flag set thinks it parsed
		t.Logf("Flag set args: %v", testFlagSet.Args())
		t.Logf("Flag set NFlag: %d", testFlagSet.NFlag())

		// Check which flags were set
		setFlags := getSetFlagsFromFlagSet(testFlagSet)

		// Debug: print all set flags
		for name, set := range setFlags {
			t.Logf("Flag '%s' is set: %v", name, set)
		}

		// Debug: print the actual flag values
		t.Logf("webhook-enabled flag value: %v", *flags.webhookEnabled)
		t.Logf("webhook-url flag value: %v", *flags.webhookURL)

		// Verify flags were parsed
		if !setFlags["webhook-enabled"] {
			t.Error("webhook-enabled flag should be set")
		}
		if !setFlags["webhook-url"] {
			t.Error("webhook-url flag should be set")
		}

		// Verify the flag values
		if *flags.webhookEnabled != true {
			t.Errorf("Expected webhook-enabled to be true, got %v", *flags.webhookEnabled)
		}
		if *flags.webhookURL != "http://test.com" {
			t.Errorf("Expected webhook-url to be 'http://test.com', got %v", *flags.webhookURL)
		}
	})
}
