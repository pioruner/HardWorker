# RP40 Calculation Module

Go module for recalculating gas permeability from RP40 pressure-falloff measurements.

Module path:

```go
module github.com/pioruner/gcalc
```

Primary package:

```go
import calc "github.com/pioruner/gcalc/pkg/RP40"
```

## Purpose

The module evaluates one pressure-falloff run in two physical branches and one display branch:

- `full`: full transient / Forchheimer-style fit
- `B-22`: Darcy / Klinkenberg approximation from RP40
- `display`: a convenience recommendation for UI usage, derived from the two branches

The display recommendation is an engineering layer, not a third independent model.

## Current Practical Position

This repository is now intentionally kept as a reusable module, not as an archive of local lab artifacts. The working assumptions that should survive into future work are:

- the real calibrated manifold-volume stages are `17.3 ml`, `233 ml`, and `1445 ml`
- the remaining problem is not volume calibration itself, but identifying which calibrated stage was active for a specific run
- `B-22` is currently the safer branch for interpreting `b`
- `full` can still be useful for `K` on some high-permeability / large-`VT` runs
- `hybrid` means `K` from `full` and `b` from `B-22`
- `auto` is a post-assessment: both branches are computed first, then the module recommends which result to display

## Method Summary

### `full`

This is the current transient / Forchheimer-style implementation. It can give useful `K`, but in a number of runs it still tends to collapse `b` toward zero.

Use it carefully:

- more promising for higher-permeability runs
- especially relevant when `B-22` linearity is weak
- not yet trustworthy as a general source of physical `b`

### `B-22`

This is the Darcy-like RP40 branch. In RP40 terms it is appropriate when the transformed response behaves linearly with mean pressure.

In practice:

- it is the main working branch for low and moderate permeability
- it usually gives the more believable `b`
- its main quality signal is the linearity of `ycn*zmn` versus `Pmn`

### `display`

`display` is the value pair intended for UI use:

- `display_method = b22`: show `K` and `b` from `B-22`
- `display_method = full`: show `K` and `b` from `full`
- `display_method = hybrid`: show `K` from `full`, `b` from `B-22`
- `display_method = review`: no clean winner, inspect the run manually

## Heuristic Used By `auto`

The current recommendation logic is intentionally simple and explicit:

- if `B-22` linearity is acceptable and `full` collapses `b`, prefer `b22`
- if permeability is high, `full` collapses `b`, and the `K` gap between branches is moderate, prefer `hybrid`
- if `B-22` linearity is weak but `full` does not obviously collapse `b`, prefer `full`
- otherwise return `review`

Current internal quality hints:

- `B-22 valid`: `R2 >= 0.88`
- `B-22 marginal`: `0.82 <= R2 < 0.88`
- `full b collapsed`: full-fit `b` is very small relative to the RP40 seed
- `high perm`: max branch `K >= 100 mD`

These thresholds are practical heuristics, not formal RP40 limits.

## Physical Interpretation Notes

### What B-22 linearity means

If the `B-22` plot is close to linear, the run behaves more like Darcy flow with slip correction and less like a flow regime dominated by inertial effects.

Linearity depends not only on permeability. It also depends on:

- active reservoir volume `VT`
- starting pressure
- falloff speed
- which part of the curve is actually usable after preprocessing

That means the same sample can look more or less suitable for `B-22` depending on the measurement regime.

### Why `VT` matters so much

For this setup, `VT` affects not only the duration of the run but also how analyzable the curve is.

Practical observation:

- small `VT` can produce a short, aggressive pressure drop
- larger `VT` often stretches the run and exposes a cleaner late-time region
- that can improve `B-22` linearity even if the initial pressure is the same

So when a run looks too aggressive, the first suspicion should often be an unfavorable `VT` regime, not immediately “this sample requires full fit”.

### Low-permeability mode

`LowPerm` is present as a practical preprocessing switch, but it is not yet a faithful reconstruction of the instrument mode. Right now it changes filtering defaults only.

## Reference Observations Preserved From The Analysis

These are the main conclusions worth keeping even after removing the bulky benchmark files.

### Sample `3-18-10`

- best practical interpretation: `B-22`
- effective `VT`: about `17.3 ml`
- `K` is close to passport and repeatable across three runs

### Sample `10-22-100`

- best practical interpretation: `B-22`
- effective `VT`: about `233 ml`
- one of the cleanest and most stable cases

### Sample `12-31-100`

- borderline sample
- short files behave more like `VT = 233 ml`
- long files behave more like `VT = 1445 ml`
- on long files `full` is often better for `K`
- on the same long files `B-22` is more believable for `b`
- this is the clearest current use case for `hybrid`

## Viscosity Handling

Viscosity was already improved compared with the original extracted code:

- it is pressure-aware
- the representative pressure now comes from mean processed `Pmn`, not from a naive midpoint

Still important:

- this is closer to RP40 than before, but still not a full tabular interpolation workflow
- improving viscosity further may help result stability even when the absolute change in final `K` is modest

## Public API

Primary entry point:

```go
result, err := calc.Analyze(req)
```

### `AnalyzeRequest`

You can pass either a file or raw arrays.

```go
type AnalyzeRequest struct {
	DataFile          string
	Time              []float64
	Pressure          []float64
	DiameterMM        float64
	LengthMM          float64
	ReservoirVolumeML float64
	AtmosphericKPa    float64
	TemperatureC      float64
	PorosityPct       float64
	Gas               string
	LowPerm           bool
	MinDeltaP         float64
	MinDeltaT         float64
	StartAt           float64
}
```

Units:

- input pressure array: `MPa` gauge
- input time array: `s`
- diameter and length: `mm`
- volume: `ml`
- atmosphere: `kPa`
- temperature: `C`
- porosity: `%`

### `AnalyzeResult`

Main fields:

```go
type AnalyzeResult struct {
	ProcessedRows      int
	ProcessedTimeSec   []float64
	ProcessedPressureM []float64
	Full               FitResult
	B22                DarcyLikeResult
	Assessment         MethodAssessment
	DisplayKMD         float64
	DisplayBPsi        float64
	DisplayMethod      string
	DisplayWarnings    []string
}
```

Useful UI interpretation:

- `DisplayKMD`, `DisplayBPsi`, `DisplayMethod` are ready-to-show fields
- `Full` and `B22` expose both physical branches directly
- `Assessment` explains why the display branch was chosen

### `Full`

```go
type FitResult struct {
	K    float64
	BPa  float64
	BPsi float64
	SE   float64
	R2   float64
	A1   float64
	A2   float64
}
```

### `B-22`

```go
type DarcyLikeResult struct {
	K           float64
	B           float64
	R2          float64
	Slope       float64
	Intercept   float64
	Linearity   []B22LinearityPoint
	MaxResidual float64
}
```

Linearity points for plotting:

```go
type B22LinearityPoint struct {
	MeanPressurePa  float64
	MeanPressurePsi float64
	Response        float64
	FitResponse     float64
	Residual        float64
}
```

### Assessment

```go
type MethodAssessment struct {
	Full            FitResult
	Darcy           DarcyLikeResult
	SeedBPsi        float64
	Rows            int
	Recommendation  string
	KSource         string
	BSource         string
	Rationale       []string
	B22Valid        bool
	B22Marginal     bool
	FullBCollapsed  bool
	HighPerm        bool
	RelativeKGap    float64
	MeanPmPsi       float64
	RecommendedKMD  float64
	RecommendedBPsi float64
}
```

## Minimal Example

```go
package main

import (
	"fmt"

	calc "github.com/pioruner/gcalc/pkg/RP40"
)

func main() {
	result, err := calc.Analyze(calc.AnalyzeRequest{
		DataFile:          "/path/to/run.tsv",
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

	fmt.Println("Display:", result.DisplayMethod, result.DisplayKMD, result.DisplayBPsi)
	fmt.Println("Full:", result.Full.K, result.Full.BPsi)
	fmt.Println("B22:", result.B22.K, result.B22.B)
	fmt.Println("Why:", result.Assessment.Rationale)
}
```

## UI Integration Notes

For a separate UI application, the most useful return groups are:

- ready result: `DisplayKMD`, `DisplayBPsi`, `DisplayMethod`, `DisplayWarnings`
- full branch: `Full.K`, `Full.BPsi`
- `B-22` branch: `B22.K`, `B22.B`, `B22.R2`
- quality explanation: `Assessment.Rationale`, `Assessment.KSource`, `Assessment.BSource`
- linearity plot: `B22.Linearity`

Suggested plots:

- pressure falloff: `ProcessedTimeSec` vs `ProcessedPressureM`
- `B-22` linearity: `MeanPressurePsi` vs `Response`
- fit overlay: `MeanPressurePsi` vs `FitResponse`
- residual view: `MeanPressurePsi` vs `Residual`

## CLI

The repository also includes a small CLI wrapper in [main.go](/Users/cim/Documents/Projects/RP40/main.go).

Methods:

- `-method auto`
- `-method full`
- `-method b22`

Example:

```bash
go run . \
  -data /path/to/run.tsv \
  -method auto \
  -vt 233 \
  -gas he \
  -d 29.957 \
  -l 30.145 \
  -patm 100.3113 \
  -temp 26.8646 \
  -pori 22.1930
```

Optional preprocessing overrides:

- `-min-dp`
- `-min-dt`
- `-start-frac`

## Using From Another Project

Fetch the module:

```bash
go get github.com/pioruner/gcalc@main
```

Then import:

```go
import calc "github.com/pioruner/gcalc/pkg/RP40"
```

For local development:

```go
require github.com/pioruner/gcalc v0.0.0

replace github.com/pioruner/gcalc => /absolute/path/to/RP40
```

## Known Limits

- active `VT` still has to be supplied from measurement context
- `LowPerm` is only an approximation
- `full` is still not a robust source of `b` in many runs
- `B-22` should be trusted only when its linearity is acceptable
- `hybrid` is an engineering compromise, not a pure single-branch RP40 output
