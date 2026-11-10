package tide

import (
	"math"
	"time"
)

// Window is a span of time when the water is below a level.
type Window struct {
	Start, End time.Time
	Low        Extreme // lowest extreme inside the span
}

// Below returns the spans where the water is under level. Extremes must be
// in time order. Heights between extremes follow a half cosine, which is
// how tide tables are usually interpolated. A span that runs off either
// end of es is cut at the first or last extreme.
func Below(es []Extreme, level float64) []Window {
	var out []Window
	var cur *Window
	open := func(t time.Time) { cur = &Window{Start: t} }
	for i, e := range es {
		if i > 0 {
			prev := es[i-1]
			switch {
			case prev.Height >= level && e.Height < level:
				open(crossing(prev, e, level))
			case prev.Height < level && e.Height >= level:
				cur.End = crossing(prev, e, level)
				out = append(out, *cur)
				cur = nil
			}
		} else if e.Height < level {
			open(e.Time)
		}
		if cur != nil && e.Height < level && (cur.Low.Time.IsZero() || e.Height < cur.Low.Height) {
			cur.Low = e
		}
	}
	if cur != nil {
		cur.End = es[len(es)-1].Time
		out = append(out, *cur)
	}
	return out
}

func crossing(a, b Extreme, level float64) time.Time {
	c := (2*level - a.Height - b.Height) / (a.Height - b.Height)
	f := math.Acos(math.Max(-1, math.Min(1, c))) / math.Pi
	return a.Time.Add(time.Duration(f * float64(b.Time.Sub(a.Time))))
}
