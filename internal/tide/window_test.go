package tide

import (
	"testing"
	"time"
)

func TestBelow(t *testing.T) {
	at := func(h, m int) time.Time { return time.Date(2026, 11, 9, h, m, 0, 0, time.UTC) }
	es := []Extreme{
		{Time: at(0, 0), Height: 2.0, Kind: High},
		{Time: at(6, 0), Height: 0.0, Kind: Low},
		{Time: at(12, 0), Height: 1.6, Kind: High},
		{Time: at(18, 0), Height: 0.6, Kind: Low},
	}
	tests := []struct {
		name  string
		level float64
		want  [][2]time.Time
	}{
		{"midway", 1.0, [][2]time.Time{{at(3, 0), at(9, 28)}, {at(15, 23), at(18, 0)}}},
		{"none", -0.1, nil},
		{"at low", 0.0, nil},
		{"above highs", 3.0, [][2]time.Time{{at(0, 0), at(18, 0)}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Below(es, tt.level)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d windows, want %d: %v", len(got), len(tt.want), got)
			}
			for i, w := range got {
				start, end := w.Start.Truncate(time.Minute), w.End.Truncate(time.Minute)
				if !start.Equal(tt.want[i][0]) || !end.Equal(tt.want[i][1]) {
					t.Errorf("window %d: %v to %v, want %v to %v", i, w.Start, w.End, tt.want[i][0], tt.want[i][1])
				}
				if w.Low.Kind != Low {
					t.Errorf("window %d: low is %v", i, w.Low)
				}
			}
		})
	}
	if w := Below(es, 3.0); w[0].Low.Height != 0.0 {
		t.Errorf("lowest in single window: %v", w[0].Low)
	}
}
