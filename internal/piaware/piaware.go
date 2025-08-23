package piaware

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

// Aircraft represents an aircraft entry from piaware.
type Aircraft struct {
	Hex     string  `json:"hex"`
	Flight  string  `json:"flight"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	AltBaro int     `json:"alt_baro"`
}

// Data represents the piaware aircraft JSON response.
type Data struct {
	Now      float64    `json:"now"`
	Aircraft []Aircraft `json:"aircraft"`
}

// Fetch retrieves aircraft data from the given URL.
func Fetch(url string) ([]Aircraft, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %s: %s", resp.Status, string(b))
	}
	var data Data
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.Aircraft, nil
}

// NearbyAircraft is an aircraft with associated distance from the base.
type NearbyAircraft struct {
	Aircraft
	DistanceKm float64
}

// FilterAircraft returns aircraft within the radius (km) and below altitude.
func FilterAircraft(aircraft []Aircraft, baseLat, baseLon, radiusKm float64, altMax int) []NearbyAircraft {
	var result []NearbyAircraft
	for _, a := range aircraft {
		if a.Lat == 0 && a.Lon == 0 {
			continue
		}
		dist := distance(baseLat, baseLon, a.Lat, a.Lon)
		if dist <= radiusKm && a.AltBaro <= altMax {
			result = append(result, NearbyAircraft{Aircraft: a, DistanceKm: dist})
		}
	}
	return result
}

// distance calculates the haversine distance in kilometers between two points.
func distance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth radius in kilometers
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
