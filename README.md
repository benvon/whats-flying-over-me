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
  "ScrapeInterval": "1m",
  "RadiusKm": 25,
  "AltitudeMax": 10000,
  "BaseLat": 37.6213,
  "BaseLon": -122.3790,
  "DataURL": "http://localhost:8080/data/aircraft.json",
  "Notifier": {
    "Method": "email",
    "Email": {
      "SMTPServer": "smtp.example.com",
      "SMTPPort": 587,
      "Username": "user",
      "Password": "pass",
      "From": "you@example.com",
      "To": "notify@example.com"
    }
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
- `WFO_NOTIFY`
- `WFO_SMTP_SERVER`
- `WFO_SMTP_PORT`
- `WFO_SMTP_USER`
- `WFO_SMTP_PASS`
- `WFO_EMAIL_FROM`
- `WFO_EMAIL_TO`

#### Command line flags

Command line flags override all other sources and match the names above:

- `-interval` scrape interval (e.g. `30s`, `1m`)
- `-radius` radius of interest in kilometers
- `-altitude` altitude ceiling in feet
- `-lat` base latitude
- `-lon` base longitude
- `-url` data retrieval URL
- `-notify` notification method (`email`)
- `-smtp-server` SMTP server for email notifications
- `-smtp-port` SMTP port
- `-smtp-user` SMTP username
- `-smtp-pass` SMTP password
- `-email-from` sender address
- `-email-to` recipient address

### Example

```
go run ./cmd/whats-flying-over-me \
  -lat 37.6213 -lon -122.3790 \
  -radius 20 -altitude 10000 \
  -smtp-server smtp.example.com -smtp-port 587 \
  -smtp-user user -smtp-pass pass \
  -email-from you@example.com -email-to notify@example.com
```

