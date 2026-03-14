package RP40

import (
	"fmt"
	"strings"

	"github.com/pioruner/gcalc/pkg/loger"
)

// AnalyzeRequest describes one pressure-falloff run.
// Pressures are expected in MPa gauge, time in seconds, and dimensions in the same
// units as the existing CLI: mm / ml / kPa / C / percent.
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

// AnalyzeResult contains both calculation branches and the current heuristic recommendation.
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

// Analyze loads and preprocesses a run, then evaluates both calculation branches.
// If DataFile is set it takes precedence over inline Time/Pressure arrays.
func Analyze(req AnalyzeRequest) (AnalyzeResult, error) {
	var result AnalyzeResult

	logs := loger.NewDiscard()

	var (
		timeVals     []float64
		pressureVals []float64
		err          error
	)
	if req.DataFile != "" {
		timeVals, pressureVals, err = LoadData(req.DataFile, logs, false)
		if err != nil {
			return result, err
		}
	} else {
		timeVals = append([]float64(nil), req.Time...)
		pressureVals = append([]float64(nil), req.Pressure...)
	}

	if len(timeVals) == 0 || len(pressureVals) == 0 {
		return result, fmt.Errorf("missing pressure-falloff data")
	}
	if len(timeVals) != len(pressureVals) {
		return result, fmt.Errorf("time and pressure length mismatch: %d vs %d", len(timeVals), len(pressureVals))
	}

	sample := SampleData{
		Time:      timeVals,
		Pressure:  pressureVals,
		D:         req.DiameterMM,
		L:         req.LengthMM,
		Vt:        req.ReservoirVolumeML,
		Patm:      req.AtmosphericKPa,
		Temp:      req.TemperatureC,
		Pori:      req.PorosityPct,
		GasType:   parseGas(req.Gas),
		LowPerm:   req.LowPerm,
		MinDeltaP: req.MinDeltaP,
		MinDeltaT: req.MinDeltaT,
		StartAt:   req.StartAt,
	}
	if err := sample.SetLog(logs); err != nil {
		return result, err
	}
	if err := sample.ToSI(); err != nil {
		return result, err
	}
	if err := sample.Check(); err != nil {
		return result, err
	}

	assessment := sample.AssessMethod()
	result.ProcessedRows = len(sample.Time)
	result.ProcessedTimeSec = append([]float64(nil), sample.Time...)
	result.ProcessedPressureM = make([]float64, len(sample.Pressure))
	for i, pressure := range sample.Pressure {
		result.ProcessedPressureM[i] = pressure / 1_000_000
	}
	result.Full = assessment.Full
	result.B22 = assessment.Darcy
	result.Assessment = assessment
	result.DisplayKMD = assessment.RecommendedKMD
	result.DisplayBPsi = assessment.RecommendedBPsi
	result.DisplayMethod = assessment.Recommendation
	result.DisplayWarnings = makeWarnings(assessment)
	return result, nil
}

func parseGas(gas string) uint {
	switch strings.ToLower(strings.TrimSpace(gas)) {
	case "n2", "n", "nitrogen", "азот":
		return uint(N)
	default:
		return uint(He)
	}
}

func makeWarnings(assessment MethodAssessment) []string {
	var warnings []string
	if assessment.B22Marginal {
		warnings = append(warnings, "B-22 linearity is marginal; inspect the high-pressure end of the regression.")
	}
	if !assessment.B22Valid && !assessment.B22Marginal {
		warnings = append(warnings, "B-22 linearity is weak; Darcy-like interpretation may be unreliable.")
	}
	if assessment.FullBCollapsed {
		warnings = append(warnings, "Full-fit b collapsed relative to the RP40 seed; treat full-fit b with caution.")
	}
	if assessment.Recommendation == "review" {
		warnings = append(warnings, "No clear display recommendation; review VT, preprocessing window, and both calculation branches.")
	}
	return warnings
}
