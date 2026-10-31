# tidemark

Command-line tide times from NOAA CO-OPS and the Brazilian Navy (DHN) station tables. Predictions are cached locally. Filter low and high water by daylight and height, or export the windows as an .ics file.

## Build

    go build ./cmd/tidemark
    go test ./...

Requires Go 1.22 or later.

## Status

The data model, the `Source` interface and the NOAA source exist so far. The commands (`stations`, `today`, `week`, `windows`, `ics`) are planned.
