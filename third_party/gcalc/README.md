# RP40 Calculation Module

Go module for recalculating gas permeability from RP40 pressure-falloff measurements.

The repository currently contains:

- a reusable calculation package: [pkg/RP40](/Users/cim/Documents/Projects/RP40/pkg/RP40)
- a CLI wrapper for local diagnostics: [main.go](/Users/cim/Documents/Projects/RP40/main.go)
- benchmark and analysis notes for the known reference samples

## What The Module Does

The package evaluates one processed pressure-falloff run in three ways:

- `full`: full transient / Forchheimer fit
- `B-22`: Darcy / Klinkenberg approximation from RP40
- `display`: convenience output for a UI, derived from the current heuristic recommendation of `b22`, `full`, `hybrid`, or `review`

The heuristic is intentionally transparent:

- it is not a pure RP40 doctrine;
- it is an engineering layer on top of the two calculation branches;
- it should be treated as improvable as more benchmark runs are collected.

## Public API

Primary entry point:

- `RP40.Analyze(req AnalyzeRequest) (AnalyzeResult, error)`

Request type:

- `AnalyzeRequest` accepts either:
  - `DataFile` with a path to the tab-separated export
  - or raw `Time` / `Pressure` arrays
- metadata matches the current CLI:
  - `DiameterMM`
  - `LengthMM`
  - `ReservoirVolumeML`
  - `AtmosphericKPa`
  - `TemperatureC`
  - `PorosityPct`
  - `Gas`
  - `LowPerm`
  - optional preprocessing overrides: `MinDeltaP`, `MinDeltaT`, `StartAt`

Result type:

- `AnalyzeResult.Full`
- `AnalyzeResult.B22`
- `AnalyzeResult.Assessment`
- `AnalyzeResult.DisplayKMD`
- `AnalyzeResult.DisplayBPsi`
- `AnalyzeResult.DisplayMethod`
- `AnalyzeResult.DisplayWarnings`
- `AnalyzeResult.ProcessedRows`
- `AnalyzeResult.ProcessedTimeSec`
- `AnalyzeResult.ProcessedPressureM`

Assessment fields that are usually the most useful for a UI:

- `Recommendation`
- `KSource`
- `BSource`
- `RecommendedKMD`
- `RecommendedBPsi`
- `Rationale`

For plotting B-22 linearity:

- `AnalyzeResult.B22.Linearity`
  - `MeanPressurePa`
  - `MeanPressurePsi`
  - `Response`
  - `FitResponse`
  - `Residual`
- `AnalyzeResult.B22.Slope`
- `AnalyzeResult.B22.Intercept`
- `AnalyzeResult.B22.R2`
- `AnalyzeResult.B22.MaxResidual`

## Example

```go
package main

import (
	"fmt"
	calc "github.com/pioruner/gcalc/pkg/RP40"
)

func main() {
	result, err := calc.Analyze(calc.AnalyzeRequest{
		DataFile:          "samples/10-22-100 - 2024-02-27_10_52_02.xls",
		DiameterMM:        29.957,
		LengthMM:          30.145,
		ReservoirVolumeML: 233,
		AtmosphericKPa:    100.3113,
		TemperatureC:      26.8646,
		PorosityPct:       22.1930,
		Gas:               "he",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Display K:", result.DisplayKMD)
	fmt.Println("Display b:", result.DisplayBPsi)
	fmt.Println("Display method:", result.DisplayMethod)
	fmt.Println("Display warnings:", result.DisplayWarnings)

	fmt.Println("Full K:", result.Full.K, "Full b:", result.Full.BPsi)
	fmt.Println("B22 K:", result.B22.K, "B22 b:", result.B22.B)
	fmt.Println("Recommendation:", result.Assessment.Recommendation)
	fmt.Println("Recommended K:", result.Assessment.RecommendedKMD, "from", result.Assessment.KSource)
	fmt.Println("Recommended b:", result.Assessment.RecommendedBPsi, "from", result.Assessment.BSource)
	fmt.Println("B22 linearity points:", len(result.B22.Linearity))
}
```

## Using From Another Project

Current `go.mod`:

```go
module github.com/pioruner/gcalc
```

Import from another project like this:

```go
import calc "github.com/pioruner/gcalc/pkg/RP40"
```

Example `go.mod` in the UI project:

```go
require github.com/pioruner/gcalc v0.0.0-<commit-or-tag>
```

If you want to work against a local checkout during development, use:

```go
require github.com/pioruner/gcalc v0.0.0

replace github.com/pioruner/gcalc => /absolute/path/to/RP40
```

## CLI

The CLI is still useful for manual checks:

- `-method full`
- `-method b22`
- `-method auto`

Example:

```bash
go run . -data "samples/10-22-100 - 2024-02-27_10_52_02.xls" -method auto -vt 233 -gas he -d 29.957 -l 30.145 -patm 100.3113 -temp 26.8646 -pori 22.1930
```

## Current Caveats

- `VT` selection still depends on knowing which real calibrated volume stage was active for the run.
- `LowPerm` is still only a practical approximation of the device behavior.
- `DisplayKMD` / `DisplayBPsi` are convenience values for a UI. They are derived from the current heuristic recommendation and should not be confused with a separate third physical calculation branch.
- `hybrid` means:
  - `K` is recommended from `full`
  - `b` is recommended from `B-22`
  - this is an engineering interpretation, not a single pure RP40 branch.

## Benchmarks

The current reference context is summarized in:

- [HANDOFF.md](/Users/cim/Documents/Projects/RP40/HANDOFF.md)
- [PASSPORT_BENCHMARK_ALL.md](/Users/cim/Documents/Projects/RP40/PASSPORT_BENCHMARK_ALL.md)
- [SAMPLE_12_31_100_REPORT.md](/Users/cim/Documents/Projects/RP40/SAMPLE_12_31_100_REPORT.md)
