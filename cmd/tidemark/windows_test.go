package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gustavosilva/tidemark/internal/tide"
)

func TestPrintWindows(t *testing.T) {
	// Seattle is UTC-8 in November.
	at := func(d, h, m int) time.Time { return time.Date(2026, 11, d, h, m, 0, 0, time.UTC) }
	es := []tide.Extreme{
		{Time: at(9, 8, 0), Height: 2.0, Kind: tide.High},
		{Time: at(9, 14, 0), Height: 0.0, Kind: tide.Low},  // 06:00 local
		{Time: at(9, 20, 0), Height: 2.0, Kind: tide.High}, // 12:00
		{Time: at(10, 2, 0), Height: 0.0, Kind: tide.Low},  // 18:00
		{Time: at(10, 8, 0), Height: 2.0, Kind: tide.High}, // 00:00 on the 10th
		{Time: at(10, 14, 0), Height: 0.4, Kind: tide.Low},
		{Time: at(10, 20, 0), Height: 2.0, Kind: tide.High},
	}
	now := at(9, 17, 0)
	tests := []struct {
		name string
		opts windowOpts
		want []string
		not  []string
	}{
		{
			name: "all day",
			opts: windowOpts{below: 1.0, days: 2, after: "00:00", before: "24:00"},
			want: []string{"Mon 09 Nov  03:00-09:00   0.00 m at Mon 06:00", "Mon 09 Nov  15:00-21:00", "Tue 10 Nov  03:28-08:31   0.40 m at Tue 06:00"},
		},
		{
			name: "morning only",
			opts: windowOpts{below: 1.0, days: 2, after: "05:00", before: "10:00"},
			want: []string{"Mon 09 Nov  05:00-09:00", "Tue 10 Nov  05:00-08:31"},
			not:  []string{"15:00"},
		},
		{
			name: "low threshold",
			opts: windowOpts{below: 0.3, days: 2, after: "00:00", before: "24:00"},
			want: []string{"Mon 09 Nov  04:28-07:31", "Mon 09 Nov  16:28-19:31"},
			not:  []string{"Tue 10 Nov"},
		},
		{
			name: "one day",
			opts: windowOpts{below: 1.0, days: 1, after: "00:00", before: "24:00"},
			not:  []string{"Tue 10 Nov"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeSource{es: es}
			var buf bytes.Buffer
			if err := printWindows(context.Background(), &buf, f, "9447130", now, tt.opts); err != nil {
				t.Fatal(err)
			}
			out := buf.String()
			for _, s := range tt.want {
				if !strings.Contains(out, s) {
					t.Errorf("output missing %q:\n%s", s, out)
				}
			}
			for _, s := range tt.not {
				if strings.Contains(out, s) {
					t.Errorf("output has %q:\n%s", s, out)
				}
			}
		})
	}
}

func TestPrintWindowsBadFlags(t *testing.T) {
	tests := []windowOpts{
		{below: 1, days: 1, after: "10:00", before: "09:00"},
		{below: 1, days: 1, after: "25:00", before: "24:00"},
		{below: 1, days: 1, after: "noon", before: "24:00"},
		{below: 1, days: 0, after: "00:00", before: "24:00"},
	}
	for _, o := range tests {
		if err := printWindows(context.Background(), &bytes.Buffer{}, &fakeSource{}, "9447130", time.Now(), o); err == nil {
			t.Errorf("%+v: expected error", o)
		}
	}
}
