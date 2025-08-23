package main

import (
	"fmt"
	"time"

	"github.com/example/whats-flying-over-me/internal/config"
	"github.com/example/whats-flying-over-me/internal/logger"
	"github.com/example/whats-flying-over-me/internal/notifier"
	"github.com/example/whats-flying-over-me/internal/piaware"
)

// AircraftFetcher defines the interface for fetching aircraft data.
type AircraftFetcher func(url string) ([]piaware.Aircraft, error)

func main() {
	cfg := config.Load()

	n, err := notifier.New(cfg.Notifier)
	if err != nil {
		logger.Critical("failed to initialize notifier", map[string]interface{}{"error": err.Error()})
		return
	}

	deduplicator := notifier.NewDeduplicator(cfg.AlertDedupe)
	stats := notifier.NewStats()

	// Create monitoring service
	monitorService := NewMonitorService(cfg, n, deduplicator, stats, piaware.Fetch)

	ticker := time.NewTicker(cfg.ScrapeInterval)
	defer ticker.Stop()

	// Start heartbeat ticker (every 5 minutes)
	heartbeatTicker := time.NewTicker(5 * time.Minute)
	defer heartbeatTicker.Stop()

	// Log startup configuration
	logger.Warn("starting aircraft monitoring", map[string]interface{}{
		"scrape_interval":  cfg.ScrapeInterval.String(),
		"radius_km":        cfg.RadiusKm,
		"altitude_max":     cfg.AltitudeMax,
		"base_lat":         cfg.BaseLat,
		"base_lon":         cfg.BaseLon,
		"data_url":         cfg.DataURL,
		"console_logging":  cfg.Notifier.Console,
		"webhook_enabled":  cfg.Notifier.Webhook.Enabled,
		"rabbitmq_enabled": cfg.Notifier.RabbitMQ.Enabled,
		"dedupe_enabled":   cfg.AlertDedupe.Enabled,
		"blockout_min":     cfg.AlertDedupe.BlockoutMin.String(),
	})

	// Start monitoring loop
	for {
		select {
		case <-ticker.C:
			if err := monitorService.RunMonitoringCycle(); err != nil {
				logger.Err("check failed", map[string]interface{}{"error": err.Error()})
				stats.RecordScrapeFailure()
			} else {
				stats.RecordScrape()
			}
		case <-heartbeatTicker.C:
			logHeartbeat(stats)
		}
	}
}

// logHeartbeat logs periodic heartbeat information about program status.
func logHeartbeat(stats *notifier.Stats) {
	statsData := stats.GetStats()

	logger.Warn("heartbeat", map[string]interface{}{
		"uptime":          statsData["uptime"],
		"scrape_count":    statsData["scrape_count"],
		"scrape_failures": statsData["scrape_failures"],
		"success_rate":    fmt.Sprintf("%.1f%%", statsData["success_rate"]),
		"unique_aircraft": statsData["unique_aircraft"],
	})
}
