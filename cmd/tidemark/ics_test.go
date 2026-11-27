package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gustavosilva/tidemark/internal/tide"
)

func TestWriteICS(t *testing.T) {
	at := func(d, h, m int) time.Time { return time.Date(2026, 11, d, h, m, 0, 0, time.UTC) }
	es := []tide.Extreme{
		{Time: at(9, 8, 0), Height: 2.0, Kind: tide.High},
		{Time: at(9, 14, 0), Height: 0.0, Kind: tide.Low},
		{Time: at(9, 20, 0), Height: 2.0, Kind: tide.High},
		{Time: at(10, 2, 0), Height: 0.0, Kind: tide.Low},
		{Time: at(10, 8, 0), Height: 2.0, Kind: tide.High},
	}
	now := at(9, 17, 0)
	tests := []struct {
		name   string
		opts   windowOpts
		events int
		want   []string
	}{
		{
			name:   "all day",
			opts:   windowOpts{below: 1.0, days: 1, after: "00:00", before: "24:00"},
			events: 2,
			want: []string{
				"UID:9447130-20261109T110000Z@tidemark\r\n",
				"DTSTAMP:20261109T170000Z\r\n",
				"DTSTART:20261109T110000Z\r\nDTEND:20261109T170000Z\r\n",
				"DTSTART:20261109T230000Z\r\nDTEND:20261110T050000Z\r\n",
				"SUMMARY:Tide below 1.00 m at Seattle\\, WA\r\n",
				"DESCRIPTION:Low 0.00 m at 20261109T140000Z\r\n",
			},
		},
		{
			name:   "morning only",
			opts:   windowOpts{below: 1.0, days: 1, after: "05:00", before: "10:00"},
			events: 1,
			want:   []string{"DTSTART:20261109T130000Z\r\nDTEND:20261109T170000Z\r\n"},
		},
		{
			name:   "none",
			opts:   windowOpts{below: -1, days: 1, after: "00:00", before: "24:00"},
			events: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := writeICS(context.Background(), &buf, &fakeSource{es: es}, "9447130", now, tt.opts); err != nil {
				t.Fatal(err)
			}
			out := buf.String()
			if !strings.HasPrefix(out, "BEGIN:VCALENDAR\r\nVERSION:2.0\r\n") || !strings.HasSuffix(out, "END:VCALENDAR\r\n") {
				t.Errorf("not a calendar:\n%s", out)
			}
			if n := strings.Count(out, "BEGIN:VEVENT\r\n"); n != tt.events {
				t.Errorf("%d events, want %d:\n%s", n, tt.events, out)
			}
			for _, s := range tt.want {
				if !strings.Contains(out, s) {
					t.Errorf("output missing %q:\n%s", s, out)
				}
			}
		})
	}
}

func TestICSText(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Seattle, WA", `Seattle\, WA`},
		{`a;b\c`, `a\;b\\c`},
		{"two\nlines", `two\nlines`},
	}
	for _, tt := range tests {
		if got := icsText.Replace(tt.in); got != tt.want {
			t.Errorf("%q: got %q, want %q", tt.in, got, tt.want)
		}
	}
}
