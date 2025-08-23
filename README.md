# whats-flying-over-me

Be alerted to what airplane is flying overhead using a local piaware

## Build and run

`whats-flying-over-me` is a Go daemon that periodically scrapes a piaware JSON feed and sends notifications when aircraft are within a configured radius and altitude of a base location.

### Configuration

Settings may be provided via a JSON config file, environment variables, or command line flags. Precedence is command line > environment > config file > defaults.

#### Config file

By default the daemon reads `config.json` in the working directory (override with the `-config` flag or `WFO_CONFIG` environment variable):

```json
{
  "scrape_interval": "1m",
  "radius_km": 25.0,
  "altitude_max": 10000,
  "base_lat": 37.6213,
  "base_lon": -122.3790,
  "data_url": "http://localhost:8080/data/aircraft.json",
  "notifier": {
    "console": true,
    "webhook": {
      "enabled": false,
      "url": "https://your-webhook-endpoint.com/aircraft-alerts",
      "timeout": "10s"
    },
    "rabbitmq": {
      "enabled": false,
      "url": "amqp://user:password@localhost:5672/",
      "exchange": "aircraft_alerts",
      "routing_key": "aircraft.nearby",
      "timeout": "10s"
    }
  },
  "alert_dedupe": {
    "enabled": true,
    "blockout_min": "15m"
  }
}
```

#### Environment variables

Each setting also has an environment variable (`WFO_*`) counterpart:

- `WFO_INTERVAL`
- `WFO_RADIUS_KM`
- `WFO_ALTITUDE_MAX`
- `WFO_BASE_LAT`
- `WFO_BASE_LON`
- `WFO_DATA_URL`

**Webhook settings:**
- `WFO_WEBHOOK_ENABLED`
- `WFO_WEBHOOK_URL`
- `WFO_WEBHOOK_TIMEOUT`

**RabbitMQ settings:**
- `WFO_RABBITMQ_ENABLED`
- `WFO_RABBITMQ_URL`
- `WFO_RABBITMQ_EXCHANGE`
- `WFO_RABBITMQ_ROUTING_KEY`
- `WFO_RABBITMQ_TIMEOUT`

**Alert deduplication settings:**
- `WFO_ALERT_DEDUPE_ENABLED`
- `WFO_ALERT_BLOCKOUT_MIN`

#### Command line flags

Command line flags override all other sources:

- `-interval` scrape interval (e.g. `30s`, `1m`)
- `-radius` radius of interest in kilometers
- `-altitude` altitude ceiling in feet
- `-lat` base latitude
- `-lon` base longitude
- `-url` data retrieval URL

**Webhook flags:**
- `-webhook-enabled` enable webhook notifications
- `-webhook-url` webhook endpoint URL
- `-webhook-timeout` webhook request timeout

**RabbitMQ flags:**
- `-rabbitmq-enabled` enable RabbitMQ notifications
- `-rabbitmq-url` RabbitMQ connection URL
- `-rabbitmq-exchange` RabbitMQ exchange name
- `-rabbitmq-routing-key` RabbitMQ routing key
- `-rabbitmq-timeout` RabbitMQ operation timeout

**Alert deduplication flags:**
- `-alert-dedupe-enabled` enable alert deduplication
- `-alert-blockout-min` alert blockout period

### Notification System

The program supports multiple notification methods that can be used simultaneously:

#### Console Logging (Default)
By default, all alerts are logged to the console in JSON format. This can be disabled by setting `console: false` in the configuration.

#### Webhook Notifications
Send alerts to HTTP webhooks. Configure with:
- `enabled`: Set to `true` to enable webhook notifications
- `url`: The webhook endpoint URL
- `timeout`: Request timeout (default: 10s)

#### RabbitMQ Messaging
Publish alerts to RabbitMQ exchanges. Configure with:
- `enabled`: Set to `true` to enable RabbitMQ notifications
- `url`: RabbitMQ connection URL (e.g., `amqp://user:pass@localhost:5672/`)
- `exchange`: Exchange name for publishing messages
- `routing_key`: Routing key for message routing
- `timeout`: Operation timeout (default: 10s)

### Alert Deduplication

To prevent notification spam, the program includes an alert deduplication system:

- **Enabled by default** with a 15-minute blockout window
- Tracks aircraft by both **tail number** and **transponder code**
- If an aircraft is seen with the same tail number but a **new transponder code** within the blockout window, a new alert will be triggered
- Configurable via `alert_dedupe.enabled` and `alert_blockout_min`

### Logging and Monitoring

The program provides comprehensive logging and monitoring to help you understand its operation:

#### Startup Logging
When the program starts, it logs its configuration including all settings and enabled features.

#### Scrape Logging
For each data scrape attempt:
- **Success**: Logs the number of aircraft seen and how many are in range
- **Failure**: Logs detailed error information for troubleshooting
- **No aircraft in range**: Logs that no aircraft meet the notification criteria

#### Heartbeat Logging
Every 5 minutes, the program logs a heartbeat message with:
- **Uptime**: How long the program has been running
- **Scrape count**: Total successful data scrapes
- **Scrape failures**: Total failed data scrapes
- **Success rate**: Percentage of successful scrapes
- **Unique aircraft**: Total unique aircraft seen since startup

#### Aircraft Tracking
The program tracks all aircraft seen (not just those in range) and provides:
- Total aircraft seen per scrape
- Aircraft in range per scrape
- Unique aircraft count over time

#### Log Format
All logs are in structured JSON format for easy parsing:
```json
{
  "timestamp": "2025-08-23T10:30:00Z",
  "level": "WARN",
  "message": "heartbeat",
  "fields": {
    "uptime": "15m30s",
    "scrape_count": 93,
    "scrape_failures": 2,
    "success_rate": "97.9%",
    "unique_aircraft": 45
  }
}
```

### Alert Data Structure

All notifications include the complete aircraft data in JSON format:

```json
{
  "timestamp": "2025-08-23T10:30:00Z",
  "aircraft": {
    "hex": "ABC123",
    "flight": "UAL123",
    "lat": 37.6213,
    "lon": -122.3790,
    "alt_baro": 5000,
    "DistanceKm": 15.2
  },
  "alert_type": "aircraft_nearby",
  "description": "Aircraft ABC123 detected within 15.2 km at 5000 ft altitude"
}
```

### Example Usage

#### Basic console logging only:
```bash
go run ./cmd/whats-flying-over-me \
  -lat 37.6213 -lon -122.3790 \
  -radius 20 -altitude 10000
```

#### With webhook notifications:
```bash
go run ./cmd/whats-flying-over-me \
  -lat 37.6213 -lon -122.3790 \
  -radius 20 -altitude 10000 \
  -webhook-enabled -webhook-url "https://your-webhook.com/alerts"
```

#### With RabbitMQ notifications:
```bash
go run ./cmd/whats-flying-over-me \
  -lat 37.6213 -lon -122.3790 \
  -radius 20 -altitude 10000 \
  -rabbitmq-enabled \
  -rabbitmq-url "amqp://user:pass@localhost:5672/" \
  -rabbitmq-exchange "aircraft_alerts" \
  -rabbitmq-routing-key "aircraft.nearby"
```

#### Custom alert deduplication:
```bash
go run ./cmd/whats-flying-over-me \
  -lat 37.6213 -lon -122.3790 \
  -radius 20 -altitude 10000 \
  -alert-blockout-min "30m"
```

### Troubleshooting

#### Common Issues

1. **No aircraft alerts**: Check that your radius and altitude settings are appropriate for your location
2. **Connection errors**: Verify the piaware data URL is accessible
3. **High failure rate**: Check network connectivity and piaware service status

#### Monitoring Program Health

Use the heartbeat logs to monitor program health:
- **Success rate below 90%**: May indicate network or service issues
- **High failure rate**: Check piaware service and network connectivity
- **No unique aircraft**: Verify piaware is receiving ADS-B data

#### Log Analysis

The structured JSON logs can be easily parsed by log aggregation tools like:
- ELK Stack (Elasticsearch, Logstash, Kibana)
- Splunk
- Grafana Loki
- Cloud logging services (AWS CloudWatch, Google Cloud Logging, etc.)

