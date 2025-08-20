package main

import (
	"fmt"
	"time"

	"github.com/example/whats-flying-over-me/internal/config"
	"github.com/example/whats-flying-over-me/internal/logger"
	"github.com/example/whats-flying-over-me/internal/notifier"
	"github.com/example/whats-flying-over-me/internal/piaware"
)

func main() {
	cfg := config.Load()

	n, err := notifier.New(cfg.Notifier)
	if err != nil {
		logger.Critical("failed to initialize notifier", map[string]interface{}{"error": err.Error()})
		return
	}

	ticker := time.NewTicker(cfg.ScrapeInterval)
	defer ticker.Stop()
	notified := make(map[string]bool)

	for {
		if err := check(cfg, n, notified); err != nil {
			logger.Err("check failed", map[string]interface{}{"error": err.Error()})
		}
		<-ticker.C
	}
}

func check(cfg config.Config, n notifier.Notifier, notified map[string]bool) error {
	aircraft, err := piaware.Fetch(cfg.DataURL)
	if err != nil {
		return err
	}
	nearby := piaware.FilterAircraft(aircraft, cfg.BaseLat, cfg.BaseLon, cfg.RadiusKm, cfg.AltitudeMax)

	for _, a := range nearby {
		if notified[a.Hex] {
			continue
		}
		subject := fmt.Sprintf("Aircraft %s nearby", a.Hex)
		body := fmt.Sprintf("Flight: %s\nAltitude: %d ft\nDistance: %.1f km", a.Flight, a.AltBaro, a.DistanceKm)
		if err := n.Notify(subject, body); err != nil {
			logger.Err("notification failed", map[string]interface{}{"hex": a.Hex, "error": err.Error()})
			continue
		}
		notified[a.Hex] = true
		logger.Warn("notified", map[string]interface{}{"hex": a.Hex, "flight": a.Flight})
	}
	return nil
}
