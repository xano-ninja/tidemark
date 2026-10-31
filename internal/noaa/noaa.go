// Package noaa implements tide.Source against the NOAA CO-OPS APIs.
package noaa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gustavosilva/tidemark/internal/tide"
)

const (
	DataURL     = "https://api.tidesandcurrents.noaa.gov/api/prod/datagetter"
	MetadataURL = "https://api.tidesandcurrents.noaa.gov/mdapi/prod/webapi"
)

type Source struct {
	Client      *http.Client
	DataURL     string
	MetadataURL string
}

var _ tide.Source = (*Source)(nil)

func New() *Source {
	return &Source{
		Client:      &http.Client{Timeout: 30 * time.Second},
		DataURL:     DataURL,
		MetadataURL: MetadataURL,
	}
}

func (s *Source) Name() string { return "noaa" }

type stationsResponse struct {
	Stations []struct {
		ID       string  `json:"id"`
		Name     string  `json:"name"`
		State    string  `json:"state"`
		Lat      float64 `json:"lat"`
		Lng      float64 `json:"lng"`
		Timezone string  `json:"timezone"`
		TZCorr   int     `json:"timezonecorr"`
	} `json:"stations"`
}

func (s *Source) Stations(ctx context.Context) ([]tide.Station, error) {
	u := s.MetadataURL + "/stations.json?type=tidepredictions"
	var r stationsResponse
	if err := s.get(ctx, u, &r); err != nil {
		return nil, err
	}
	out := make([]tide.Station, 0, len(r.Stations))
	for _, st := range r.Stations {
		name := st.Name
		if st.State != "" {
			name += ", " + st.State
		}
		out = append(out, tide.Station{
			ID:       st.ID,
			Name:     name,
			Lat:      st.Lat,
			Lon:      st.Lng,
			Timezone: zone(st.Timezone, st.TZCorr),
		})
	}
	return out, nil
}

// NOAA reports zones as US abbreviations. Unknown ones fall back to a
// fixed offset, which loses DST but keeps the hour right half the year.
var zones = map[string]string{
	"EST":  "America/New_York",
	"CST":  "America/Chicago",
	"MST":  "America/Denver",
	"PST":  "America/Los_Angeles",
	"AKST": "America/Anchorage",
	"HAST": "Pacific/Honolulu",
	"HST":  "Pacific/Honolulu",
	"AST":  "America/Puerto_Rico",
	"SST":  "Pacific/Pago_Pago",
	"ChST": "Pacific/Guam",
}

func zone(abbr string, corr int) string {
	if z, ok := zones[abbr]; ok {
		return z
	}
	switch {
	case corr == 0:
		return "UTC"
	case corr < 0:
		return fmt.Sprintf("Etc/GMT+%d", -corr)
	default:
		return fmt.Sprintf("Etc/GMT-%d", corr)
	}
}

type predictionsResponse struct {
	Predictions []struct {
		T    string `json:"t"`
		V    string `json:"v"`
		Type string `json:"type"`
	} `json:"predictions"`
}

func (s *Source) Extremes(ctx context.Context, stationID string, from, to time.Time) ([]tide.Extreme, error) {
	q := url.Values{
		"product":     {"predictions"},
		"application": {"tidemark"},
		"station":     {stationID},
		"begin_date":  {from.UTC().Format("20060102 15:04")},
		"end_date":    {to.UTC().Format("20060102 15:04")},
		"datum":       {"MLLW"},
		"time_zone":   {"gmt"},
		"interval":    {"hilo"},
		"units":       {"metric"},
		"format":      {"json"},
	}
	var r predictionsResponse
	if err := s.get(ctx, s.DataURL+"?"+q.Encode(), &r); err != nil {
		return nil, err
	}
	out := make([]tide.Extreme, 0, len(r.Predictions))
	for _, p := range r.Predictions {
		t, err := time.Parse("2006-01-02 15:04", p.T)
		if err != nil {
			return nil, fmt.Errorf("noaa: bad time %q: %w", p.T, err)
		}
		h, err := strconv.ParseFloat(p.V, 64)
		if err != nil {
			return nil, fmt.Errorf("noaa: bad height %q: %w", p.V, err)
		}
		var k tide.Kind
		switch p.Type {
		case "L":
			k = tide.Low
		case "H":
			k = tide.High
		default:
			return nil, fmt.Errorf("noaa: unknown type %q", p.Type)
		}
		out = append(out, tide.Extreme{Time: t, Height: h, Kind: k})
	}
	return out, nil
}

type apiError struct {
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (s *Source) get(ctx context.Context, u string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("noaa: %s", resp.Status)
	}
	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return fmt.Errorf("noaa: decode: %w", err)
	}
	// The data API reports errors with a 200 and an error object.
	var e apiError
	if err := json.Unmarshal(raw, &e); err == nil && e.Error != nil {
		return errors.New("noaa: " + e.Error.Message)
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("noaa: decode: %w", err)
	}
	return nil
}
