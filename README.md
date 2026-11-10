# tidemark

Command-line tide times from NOAA CO-OPS and the Brazilian Navy (DHN) station tables. Predictions are cached locally. Filter low and high water by daylight and height, or export the windows as an .ics file.

## Build

    go build ./cmd/tidemark
    go test ./...

Requires Go 1.22 or later.

## Usage

    tidemark stations seattle
    tidemark today 9447130
    tidemark week 9447130
    tidemark windows 9447130 --below 0.3 --before 10:00

`windows` lists the spans when the water is under `--below` metres, cut to the local hours between `--after` and `--before`, over `--days` days (default 7). Heights between low and high water are interpolated along a half cosine.

Times are shown in the station's local zone. Station lists and predictions are cached under the user cache directory (`~/.cache/tidemark` on Linux). Station lists are refetched after seven days; predictions are kept.

## Status

Only the NOAA source is wired up. `ics` and the DHN source are planned.
