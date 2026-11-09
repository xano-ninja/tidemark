package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gustavosilva/tidemark/internal/tide"
)

type fakeSource struct{ from, to time.Time }

func (f *fakeSource) Name() string { return "fake" }

func (f *fakeSource) Stations(context.Context) ([]tide.Station, error) {
	return []tide.Station{{ID: "9447130", Name: "Seattle, WA", Timezone: "America/Los_Angeles"}}, nil
}

func (f *fakeSource) Extremes(_ context.Context, _ string, from, to time.Time) ([]tide.Extreme, error) {
	f.from, f.to = from, to
	return []tide.Extreme{
		{Time: time.Date(2026, 11, 9, 13, 42, 0, 0, time.UTC), Height: -0.31, Kind: tide.Low},
		{Time: time.Date(2026, 11, 9, 20, 5, 0, 0, time.UTC), Height: 3.12, Kind: tide.High},
	}, nil
}

func TestPrintDays(t *testing.T) {
	// 03:00 UTC on the 10th is still the 9th in Seattle.
	now := time.Date(2026, 11, 10, 3, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		days     int
		from, to time.Time
	}{
		{"today", 1, time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC), time.Date(2026, 11, 10, 8, 0, 0, 0, time.UTC)},
		{"week", 7, time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC), time.Date(2026, 11, 16, 8, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeSource{}
			var buf bytes.Buffer
			if err := printDays(context.Background(), &buf, f, "9447130", now, tt.days); err != nil {
				t.Fatal(err)
			}
			if !f.from.Equal(tt.from) || !f.to.Equal(tt.to) {
				t.Errorf("range %v to %v, want %v to %v", f.from.UTC(), f.to.UTC(), tt.from, tt.to)
			}
			out := buf.String()
			for _, want := range []string{"Seattle, WA (9447130)", "Mon 09 Nov  05:42  low   -0.31 m", "12:05  high   3.12 m"} {
				if !strings.Contains(out, want) {
					t.Errorf("output missing %q:\n%s", want, out)
				}
			}
		})
	}
}

func TestPrintDaysUnknownStation(t *testing.T) {
	err := printDays(context.Background(), &bytes.Buffer{}, &fakeSource{}, "nope", time.Now(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}
