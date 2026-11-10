package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/gustavosilva/tidemark/internal/cache"
	"github.com/gustavosilva/tidemark/internal/noaa"
	"github.com/gustavosilva/tidemark/internal/tide"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "tidemark:", err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "tidemark",
		Short:         "Low and high water times for a tide station",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(stationsCmd(), todayCmd(), weekCmd(), windowsCmd())
	return root
}

func source() (tide.Source, error) {
	dir, err := cache.Dir()
	if err != nil {
		return nil, err
	}
	return cache.New(noaa.New(), dir), nil
}

func stationsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stations [name filter]",
		Short: "List stations",
		RunE: func(cmd *cobra.Command, args []string) error {
			src, err := source()
			if err != nil {
				return err
			}
			list, err := src.Stations(cmd.Context())
			if err != nil {
				return err
			}
			filter := strings.ToLower(strings.Join(args, " "))
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			for _, s := range list {
				if filter != "" && !strings.Contains(strings.ToLower(s.Name), filter) {
					continue
				}
				fmt.Fprintf(w, "%s\t%s\t%.4f\t%.4f\t%s\n", s.ID, s.Name, s.Lat, s.Lon, s.Timezone)
			}
			return w.Flush()
		},
	}
}
