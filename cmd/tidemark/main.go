package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/gustavosilva/tidemark/internal/noaa"
	"github.com/gustavosilva/tidemark/internal/tide"
)

const usage = "usage: tidemark <stations|today|week|windows|ics>"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "-h", "--help":
		fmt.Println(usage)
		return
	case "stations":
		err = stations(context.Background(), noaa.New(), os.Args[2:])
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "tidemark:", err)
		os.Exit(1)
	}
}

func stations(ctx context.Context, src tide.Source, args []string) error {
	list, err := src.Stations(ctx)
	if err != nil {
		return err
	}
	filter := strings.ToLower(strings.Join(args, " "))
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, s := range list {
		if filter != "" && !strings.Contains(strings.ToLower(s.Name), filter) {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%.4f\t%.4f\t%s\n", s.ID, s.Name, s.Lat, s.Lon, s.Timezone)
	}
	return w.Flush()
}
