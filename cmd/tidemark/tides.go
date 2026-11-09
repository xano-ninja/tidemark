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

func todayCmd() *cobra.Command {
	return daysCmd("today <station id>", "Show today's tides at a station", 1)
}

func weekCmd() *cobra.Command {
	return daysCmd("week <station id>", "Show tides for the next seven days", 7)
}

func daysCmd(use, short string, days int) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, err := source()
			if err != nil {
				return err
			}
			return printDays(cmd.Context(), cmd.OutOrStdout(), src, args[0], time.Now(), days)
		},
	}
}

// printDays prints extremes from the start of now's day in the station's
// zone through the given number of local days.
func printDays(ctx context.Context, w io.Writer, src tide.Source, id string, now time.Time, days int) error {
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
	to := from.AddDate(0, 0, days)
	es, err := src.Extremes(ctx, st.ID, from, to)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%s (%s)\n", st.Name, st.ID)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, e := range es {
		t := e.Time.In(loc)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%5.2f m\n", t.Format("Mon 02 Jan"), t.Format("15:04"), e.Kind, e.Height)
	}
	return tw.Flush()
}

func findStation(ctx context.Context, src tide.Source, id string) (tide.Station, error) {
	list, err := src.Stations(ctx)
	if err != nil {
		return tide.Station{}, err
	}
	for _, s := range list {
		if s.ID == id {
			return s, nil
		}
	}
	return tide.Station{}, fmt.Errorf("no station %q at %s", id, src.Name())
}
