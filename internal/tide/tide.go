// Package tide defines the data model and the Source interface that
// each prediction provider implements.
package tide

import (
	"context"
	"time"
)

// Kind is the type of a predicted extreme.
type Kind string

const (
	Low  Kind = "low"
	High Kind = "high"
)

// Station is a tide station at a provider.
type Station struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Timezone string  `json:"timezone"` // IANA name
}

// Extreme is one predicted low or high water. Time is always UTC.
type Extreme struct {
	Time   time.Time `json:"time"`
	Height float64   `json:"height"` // metres
	Kind   Kind      `json:"kind"`
}

// Local returns the extreme's time in the station's zone.
func (s Station) Local(e Extreme) (time.Time, error) {
	loc, err := time.LoadLocation(s.Timezone)
	if err != nil {
		return time.Time{}, err
	}
	return e.Time.In(loc), nil
}

// Source is a tide prediction provider.
type Source interface {
	Name() string
	Stations(ctx context.Context) ([]Station, error)
	Extremes(ctx context.Context, stationID string, from, to time.Time) ([]Extreme, error)
}
