package noaa

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gustavosilva/tidemark/internal/tide"
)

func serve(t *testing.T, fixture string, check func(url.Values)) *Source {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if check != nil {
			check(r.URL.Query())
		}
		http.ServeFile(w, r, "testdata/"+fixture)
	}))
	t.Cleanup(srv.Close)
	return &Source{Client: srv.Client(), DataURL: srv.URL + "/datagetter", MetadataURL: srv.URL}
}

func TestStations(t *testing.T) {
	s := serve(t, "stations.json", func(q url.Values) {
		if q.Get("type") != "tidepredictions" {
			t.Errorf("type = %q", q.Get("type"))
		}
	})
	got, err := s.Stations(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []tide.Station{
		{ID: "9447130", Name: "Seattle, WA", Lat: 47.6026, Lon: -122.3393, Timezone: "America/Los_Angeles"},
		{ID: "1612340", Name: "Honolulu, HI", Lat: 21.3033, Lon: -157.8645, Timezone: "Pacific/Honolulu"},
		{ID: "1631428", Name: "Pago Bay", Lat: 13.4283, Lon: 144.7967, Timezone: "Etc/GMT-10"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d stations, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("station %d: got %+v, want %+v", i, got[i], want[i])
		}
		if _, err := time.LoadLocation(got[i].Timezone); err != nil {
			t.Errorf("station %d: %v", i, err)
		}
	}
}

func TestExtremes(t *testing.T) {
	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)
	s := serve(t, "predictions.json", func(q url.Values) {
		for k, v := range map[string]string{
			"station":    "9447130",
			"begin_date": "20261101 00:00",
			"end_date":   "20261102 00:00",
			"time_zone":  "gmt",
			"interval":   "hilo",
			"units":      "metric",
		} {
			if q.Get(k) != v {
				t.Errorf("%s = %q, want %q", k, q.Get(k), v)
			}
		}
	})
	got, err := s.Extremes(context.Background(), "9447130", from, to)
	if err != nil {
		t.Fatal(err)
	}
	want := []tide.Extreme{
		{Time: time.Date(2026, 11, 1, 4, 51, 0, 0, time.UTC), Height: 3.512, Kind: tide.High},
		{Time: time.Date(2026, 11, 1, 10, 2, 0, 0, time.UTC), Height: 1.874, Kind: tide.Low},
		{Time: time.Date(2026, 11, 1, 14, 47, 0, 0, time.UTC), Height: 3.031, Kind: tide.High},
		{Time: time.Date(2026, 11, 1, 22, 20, 0, 0, time.UTC), Height: -0.214, Kind: tide.Low},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d extremes, want %d", len(got), len(want))
	}
	for i := range want {
		if !got[i].Time.Equal(want[i].Time) || got[i].Height != want[i].Height || got[i].Kind != want[i].Kind {
			t.Errorf("extreme %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestExtremesAPIError(t *testing.T) {
	s := serve(t, "error.json", nil)
	_, err := s.Extremes(context.Background(), "0000000", time.Now(), time.Now().Add(time.Hour))
	if err == nil {
		t.Fatal("expected error")
	}
	want := "noaa: No Predictions data was found. Please make sure the Datum input is valid."
	if err.Error() != want {
		t.Errorf("got %q, want %q", err, want)
	}
}

func TestZone(t *testing.T) {
	tests := []struct {
		abbr string
		corr int
		want string
	}{
		{"EST", -5, "America/New_York"},
		{"AKST", -9, "America/Anchorage"},
		{"XYZ", -4, "Etc/GMT+4"},
		{"XYZ", 10, "Etc/GMT-10"},
		{"", 0, "UTC"},
	}
	for _, tt := range tests {
		if got := zone(tt.abbr, tt.corr); got != tt.want {
			t.Errorf("zone(%q, %d) = %q, want %q", tt.abbr, tt.corr, got, tt.want)
		}
	}
}
