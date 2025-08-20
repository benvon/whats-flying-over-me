package piaware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDistance(t *testing.T) {
	if d := distance(0, 0, 0, 0); d != 0 {
		t.Fatalf("expected 0 distance, got %v", d)
	}
}

func TestFilterAircraft(t *testing.T) {
	aircraft := []Aircraft{
		{Hex: "1", Lat: 0.1, Lon: 0.1, AltBaro: 1000},
		{Hex: "2", Lat: 50, Lon: 50, AltBaro: 2000},
	}
	res := FilterAircraft(aircraft, 0, 0, 1000, 1500)
	if len(res) != 1 || res[0].Hex != "1" {
		t.Fatalf("unexpected filter result: %+v", res)
	}
}

func TestFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, `{"now":1,"aircraft":[{"hex":"abc","flight":"TEST","lat":1,"lon":2,"alt_baro":300}]}`); err != nil {
			t.Fatalf("write: %v", err)
		}
	}))
	defer srv.Close()
	ac, err := Fetch(srv.URL)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(ac) != 1 || ac[0].Hex != "abc" {
		t.Fatalf("unexpected fetch result: %+v", ac)
	}
}

func TestFetchError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	if _, err := Fetch(srv.URL); err == nil {
		t.Fatal("expected error")
	}
}
