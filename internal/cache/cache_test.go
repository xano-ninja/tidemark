package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gustavosilva/tidemark/internal/tide"
)

type fake struct {
	stations []tide.Station
	extremes []tide.Extreme
	calls    [][2]time.Time
	nst      int
}

func (f *fake) Name() string { return "fake" }

func (f *fake) Stations(context.Context) ([]tide.Station, error) {
	f.nst++
	return f.stations, nil
}

func (f *fake) Extremes(_ context.Context, _ string, from, to time.Time) ([]tide.Extreme, error) {
	f.calls = append(f.calls, [2]time.Time{from, to})
	var out []tide.Extreme
	for _, e := range f.extremes {
		if !e.Time.Before(from) && !e.Time.After(to) {
			out = append(out, e)
		}
	}
	return out, nil
}

func utc(d, h int) time.Time { return time.Date(2026, 11, d, h, 0, 0, 0, time.UTC) }

func newFake() *fake {
	f := &fake{stations: []tide.Station{{ID: "9447130", Name: "Seattle, WA", Timezone: "America/Los_Angeles"}}}
	for d := 9; d <= 12; d++ {
		for _, h := range []int{0, 6, 12, 18} {
			k := tide.Low
			if h%12 == 6 {
				k = tide.High
			}
			f.extremes = append(f.extremes, tide.Extreme{Time: utc(d, h), Height: float64(h) / 10, Kind: k})
		}
	}
	return f
}

func TestExtremes(t *testing.T) {
	tests := []struct {
		name      string
		warm      [2]time.Time
		from, to  time.Time
		want      int
		wantCalls int
	}{
		{"cold", [2]time.Time{}, utc(9, 3), utc(10, 3), 4, 1},
		{"warm", [2]time.Time{utc(9, 0), utc(11, 0)}, utc(9, 3), utc(10, 3), 4, 0},
		{"partial", [2]time.Time{utc(9, 0), utc(10, 0)}, utc(9, 0), utc(11, 0), 8, 1},
		{"end exclusive", [2]time.Time{}, utc(9, 0), utc(10, 0), 4, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFake()
			c := New(f, t.TempDir())
			ctx := context.Background()
			if !tt.warm[0].IsZero() {
				if _, err := c.Extremes(ctx, "9447130", tt.warm[0], tt.warm[1]); err != nil {
					t.Fatal(err)
				}
				f.calls = nil
			}
			got, err := c.Extremes(ctx, "9447130", tt.from, tt.to)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != tt.want {
				t.Errorf("got %d extremes, want %d", len(got), tt.want)
			}
			for i := 1; i < len(got); i++ {
				if !got[i-1].Time.Before(got[i].Time) {
					t.Errorf("not sorted at %d", i)
				}
			}
			for _, e := range got {
				if e.Time.Before(tt.from) || !e.Time.Before(tt.to) {
					t.Errorf("%v outside [%v, %v)", e.Time, tt.from, tt.to)
				}
			}
			if len(f.calls) != tt.wantCalls {
				t.Errorf("got %d source calls, want %d", len(f.calls), tt.wantCalls)
			}
		})
	}
}

func TestExtremesCorruptFile(t *testing.T) {
	f := newFake()
	dir := t.TempDir()
	c := New(f, dir)
	path := filepath.Join(dir, "fake", "9447130", "2026-11-09.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := c.Extremes(context.Background(), "9447130", utc(9, 0), utc(10, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 || len(f.calls) != 1 {
		t.Errorf("got %d extremes and %d calls, want 4 and 1", len(got), len(f.calls))
	}
}

func TestStations(t *testing.T) {
	f := newFake()
	c := New(f, t.TempDir())
	now := time.Now()
	c.now = func() time.Time { return now }
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if _, err := c.Stations(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if f.nst != 1 {
		t.Fatalf("got %d fetches, want 1", f.nst)
	}
	c.now = func() time.Time { return now.Add(StationsTTL + time.Hour) }
	list, err := c.Stations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if f.nst != 2 || len(list) != 1 || list[0].ID != "9447130" {
		t.Fatalf("got %d fetches and %v", f.nst, list)
	}
}
