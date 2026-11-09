// Package cache wraps a tide.Source and keeps its results as JSON files
// on disk.
package cache

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gustavosilva/tidemark/internal/tide"
)

// StationsTTL is how long a cached station list is used before refetching.
const StationsTTL = 7 * 24 * time.Hour

const day = 24 * time.Hour

type Source struct {
	src tide.Source
	dir string
	now func() time.Time
}

var _ tide.Source = (*Source)(nil)

// New caches src under dir/<source name>.
func New(src tide.Source, dir string) *Source {
	return &Source{src: src, dir: filepath.Join(dir, src.Name()), now: time.Now}
}

// Dir returns the default cache directory.
func Dir() (string, error) {
	d, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "tidemark"), nil
}

func (s *Source) Name() string { return s.src.Name() }

func (s *Source) Stations(ctx context.Context) ([]tide.Station, error) {
	path := filepath.Join(s.dir, "stations.json")
	if fi, err := os.Stat(path); err == nil && s.now().Sub(fi.ModTime()) < StationsTTL {
		var list []tide.Station
		if err := read(path, &list); err == nil {
			return list, nil
		}
	}
	list, err := s.src.Stations(ctx)
	if err != nil {
		return nil, err
	}
	return list, write(path, list)
}

// Extremes caches predictions per station and UTC day. Predictions for a
// past or future day do not change, so day files never expire.
func (s *Source) Extremes(ctx context.Context, stationID string, from, to time.Time) ([]tide.Extreme, error) {
	from, to = from.UTC(), to.UTC()
	var (
		out     []tide.Extreme
		missing []time.Time
	)
	for d := from.Truncate(day); d.Before(to); d = d.Add(day) {
		var es []tide.Extreme
		if err := read(s.dayPath(stationID, d), &es); err != nil {
			missing = append(missing, d)
			continue
		}
		out = append(out, es...)
	}
	if len(missing) > 0 {
		first, last := missing[0], missing[len(missing)-1].Add(day)
		fetched, err := s.src.Extremes(ctx, stationID, first, last)
		if err != nil {
			return nil, err
		}
		byDay := make(map[time.Time][]tide.Extreme, len(missing))
		for _, d := range missing {
			byDay[d] = []tide.Extreme{}
		}
		for _, e := range fetched {
			d := e.Time.UTC().Truncate(day)
			if es, ok := byDay[d]; ok {
				byDay[d] = append(es, e)
			}
		}
		for d, es := range byDay {
			if err := write(s.dayPath(stationID, d), es); err != nil {
				return nil, err
			}
			out = append(out, es...)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Time.Before(out[j].Time) })
	n := 0
	for _, e := range out {
		if !e.Time.Before(from) && e.Time.Before(to) {
			out[n] = e
			n++
		}
	}
	return out[:n], nil
}

func (s *Source) dayPath(stationID string, d time.Time) string {
	return filepath.Join(s.dir, filepath.Base(stationID), d.Format("2006-01-02")+".json")
}

func read(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func write(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), path)
}
