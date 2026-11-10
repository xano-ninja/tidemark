package main

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/gustavosilva/tidemark/internal/tide"
)

type windowOpts struct {
	below         float64
	days          int
	after, before string
}

func windowsCmd() *cobra.Command {
	var o windowOpts
	cmd := &cobra.Command{
		Use:   "windows <station id>",
		Short: "Show when the water is below a height",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, err := source()
			if err != nil {
				return err
			}
			return printWindows(cmd.Context(), cmd.OutOrStdout(), src, args[0], time.Now(), o)
		},
	}
	f := cmd.Flags()
	f.Float64Var(&o.below, "below", 0, "height in metres the water must be under")
	f.IntVar(&o.days, "days", 7, "number of local days to search")
	f.StringVar(&o.after, "after", "00:00", "earliest local time of day")
	f.StringVar(&o.before, "before", "24:00", "latest local time of day")
	cmd.MarkFlagRequired("below")
	return cmd
}

func printWindows(ctx context.Context, w io.Writer, src tide.Source, id string, now time.Time, o windowOpts) error {
	after, err := clock(o.after)
	if err != nil {
		return err
	}
	before, err := clock(o.before)
	if err != nil {
		return err
	}
	if after >= before {
		return fmt.Errorf("--after %s is not before --before %s", o.after, o.before)
	}
	if o.days < 1 {
		return fmt.Errorf("--days must be at least 1")
	}
	st, err := findStation(ctx, src, id)
	if err != nil {
		return err
	}
	loc, err := time.LoadLocation(st.Timezone)
	if err != nil {
		return err
	}
	n := now.In(loc)
	from := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, loc)
	to := from.AddDate(0, 0, o.days)
	// Pad the fetch so windows crossing the range edges get real start and end times.
	es, err := src.Extremes(ctx, st.ID, from.Add(-12*time.Hour), to.Add(12*time.Hour))
	if err != nil {
		return err
	}
	ws := tide.Below(es, o.below)

	fmt.Fprintf(w, "%s (%s), below %.2f m\n", st.Name, st.ID, o.below)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for d := 0; d < o.days; d++ {
		day := from.AddDate(0, 0, d)
		lo := time.Date(day.Year(), day.Month(), day.Day(), 0, after, 0, 0, loc)
		hi := time.Date(day.Year(), day.Month(), day.Day(), 0, before, 0, 0, loc)
		for _, win := range ws {
			s, e := maxTime(win.Start, lo), minTime(win.End, hi)
			if !s.Before(e) {
				continue
			}
			low := win.Low.Time.In(loc)
			fmt.Fprintf(tw, "%s\t%s-%s\t%5.2f m at %s\n",
				day.Format("Mon 02 Jan"), s.In(loc).Format("15:04"), endClock(e.In(loc), day),
				win.Low.Height, low.Format("Mon 15:04"))
		}
	}
	return tw.Flush()
}

// clock parses HH:MM into minutes after midnight. 24:00 is allowed.
func clock(s string) (int, error) {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil || h < 0 || m < 0 || m > 59 || h*60+m > 24*60 {
		return 0, fmt.Errorf("bad time of day %q, want HH:MM", s)
	}
	return h*60 + m, nil
}

func endClock(t, day time.Time) string {
	if t.Day() != day.Day() {
		return "24:00"
	}
	return t.Format("15:04")
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
