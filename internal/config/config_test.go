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

	// set command line flag
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = []string{os.Args[0], "-radius", "30"}

	cfg := Load()
	if cfg.RadiusKm != 30 {
		t.Fatalf("expected radius 30, got %v", cfg.RadiusKm)
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
}
