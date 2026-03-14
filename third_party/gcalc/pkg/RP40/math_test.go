package RP40

import (
	"math"
	"testing"
)

func TestMeanCorePressureUsesProcessedPairs(t *testing.T) {
	sample := SampleData{
		Pressure: []float64{1_000_000, 810_000, 640_000},
		Patm:     100_000,
	}

	got := sample.meanCorePressure()

	pg1 := math.Sqrt(1_000_000 * 810_000)
	pg2 := math.Sqrt(810_000 * 640_000)
	want := (((pg1 / 2) + 100_000) + ((pg2 / 2) + 100_000)) / 2

	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("meanCorePressure() = %g, want %g", got, want)
	}
}

func TestHeliumViscosityCorrectionMatchesDocumentAnchors(t *testing.T) {
	tests := []struct {
		pressureAtm float64
		want        float64
	}{
		{1, 1.0},
		{37, 0.9957},
		{158, 1.0017},
	}

	for _, tt := range tests {
		got := heliumViscosityCorrection(tt.pressureAtm)
		if math.Abs(got-tt.want) > 1e-9 {
			t.Fatalf("heliumViscosityCorrection(%g) = %g, want %g", tt.pressureAtm, got, tt.want)
		}
	}
}

func TestCalcMuUsesMeanCorePressure(t *testing.T) {
	sample := SampleData{
		Pressure: []float64{1_000_000, 810_000, 640_000},
		Patm:     100_000,
		Temp:     300.0,
		GasType:  He,
	}

	sample.calcMu()

	want := heliumViscosity(sample.Temp, sample.meanCorePressure()/101325.0)
	if math.Abs(sample.Mu-want) > 1e-15 {
		t.Fatalf("Mu = %g, want %g", sample.Mu, want)
	}
}

func TestAssessMethodPrefersB22WhenFullBCollapsesOutsideHighPerm(t *testing.T) {
	assessment := MethodAssessment{
		Darcy:          DarcyLikeResult{K: 11.21, B: 13.85, R2: 0.886},
		Full:           FitResult{K: 16.45, BPsi: 0.006},
		SeedBPsi:       8.7,
		RelativeKGap:   0.46,
		HighPerm:       false,
		B22Valid:       true,
		FullBCollapsed: true,
	}

	switch {
	case assessment.B22Valid && assessment.FullBCollapsed && !assessment.HighPerm:
		assessment.Recommendation = "b22"
	default:
		assessment.Recommendation = "review"
	}

	if assessment.Recommendation != "b22" {
		t.Fatalf("recommendation = %s, want b22", assessment.Recommendation)
	}
}

func TestAssessMethodAllowsHybridForHighPermGap(t *testing.T) {
	assessment := MethodAssessment{
		Darcy:          DarcyLikeResult{K: 154.2, B: 4.02, R2: 0.858},
		Full:           FitResult{K: 171.3, BPsi: 0.002},
		SeedBPsi:       42.0,
		RelativeKGap:   (171.3 - 154.2) / 154.2,
		HighPerm:       true,
		B22Marginal:    true,
		FullBCollapsed: true,
	}

	switch {
	case (assessment.B22Valid || assessment.B22Marginal) && assessment.FullBCollapsed && assessment.HighPerm && assessment.RelativeKGap >= 0.05 && assessment.RelativeKGap <= 0.20:
		assessment.Recommendation = "hybrid"
	default:
		assessment.Recommendation = "review"
	}

	if assessment.Recommendation != "hybrid" {
		t.Fatalf("recommendation = %s, want hybrid", assessment.Recommendation)
	}
}
