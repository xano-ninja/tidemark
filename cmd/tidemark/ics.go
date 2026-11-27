package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gustavosilva/tidemark/internal/tide"
)

func icsCmd() *cobra.Command {
	var o windowOpts
	cmd := &cobra.Command{
		Use:   "ics <station id>",
		Short: "Write below-height windows as an iCalendar file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, err := source()
			if err != nil {
				return err
			}
			return writeICS(cmd.Context(), cmd.OutOrStdout(), src, args[0], time.Now(), o)
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

func writeICS(ctx context.Context, w io.Writer, src tide.Source, id string, now time.Time, o windowOpts) error {
	st, _, spans, err := findWindows(ctx, src, id, now, o)
	if err != nil {
		return err
	}
	const f = "20060102T150405Z"
	fmt.Fprint(w, "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//tidemark//EN\r\n")
	for _, sp := range spans {
		fmt.Fprintf(w, "BEGIN:VEVENT\r\nUID:%s-%s@tidemark\r\nDTSTAMP:%s\r\nDTSTART:%s\r\nDTEND:%s\r\nSUMMARY:Tide below %.2f m at %s\r\nDESCRIPTION:Low %.2f m at %s\r\nEND:VEVENT\r\n",
			st.ID, sp.start.UTC().Format(f), now.UTC().Format(f), sp.start.UTC().Format(f), sp.end.UTC().Format(f),
			o.below, icsText.Replace(st.Name), sp.low.Height, sp.low.Time.UTC().Format(f))
	}
	_, err = fmt.Fprint(w, "END:VCALENDAR\r\n")
	return err
}

var icsText = strings.NewReplacer(`\`, `\\`, ";", `\;`, ",", `\,`, "\n", `\n`)
