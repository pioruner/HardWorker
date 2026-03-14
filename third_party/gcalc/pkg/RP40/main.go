package RP40

import (
	"math"

	"github.com/pioruner/gcalc/pkg/loger"
)

const (
	Pmin      = 100000  //	Pa
	Pmax      = 1241056 // Pa
	MinDP     = 1000    //	Pa
	MinDT     = 0.5     //	Seconds
	StartFrac = 0.85    //	Start analysis after pressure profile stabilizes
	EndFrac   = 0.10    //	Drop tail where pressure is too low for stable sensing
	FiFoStart = 0.95    //	const
	He        = 0       //	code
	N         = 1       //	code
	//Krest     = 18.53
	//Shlang    = 238.31
	E1 = 1450.0
)

type SampleData struct {
	Time      []float64 //	Seconds
	Pressure  []float64 //	MPa
	D         float64   //	mm
	L         float64   //	mm
	Vt        float64   //	ml
	FiFo      []float64 //	const
	Patm      float64   //	Pa
	Temp      float64   //	C
	Pori      float64   //	%
	GasType   uint      //	0-1 He-N
	LowPerm   bool
	MinDeltaP float64
	MinDeltaT float64
	StartAt   float64
	Mu        float64
	Vp        float64
	Gamma     float64
	Pg        []float64 //	Pa
	Pn        []float64 //	Pa
	Yn        []float64 //	Pa
	Zn        []float64
	Pm        []float64
	Zm        []float64
	Fz        []float64
	Cn        []float64
	G         []float64
	Yc        []float64
	Xi        []float64
	Yi        []float64
	Nf0       []float64
	E         []float64
	A1        float64
	A2        float64
	b         float64
	r2        float64
	se        float64
	Logs      *loger.Logger
}

type DarcyLikeResult struct {
	K           float64
	B           float64
	R2          float64
	Slope       float64
	Intercept   float64
	Linearity   []B22LinearityPoint
	MaxResidual float64
}

type FitResult struct {
	K    float64
	BPa  float64
	BPsi float64
	SE   float64
	R2   float64
	A1   float64
	A2   float64
}

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

type B22LinearityPoint struct {
	MeanPressurePa  float64
	MeanPressurePsi float64
	Response        float64
	FitResponse     float64
	Residual        float64
}

func (data *SampleData) Calc(logging bool) (K float64, b float64) {
	K, b = 0, 0
	if data.Logs == nil {
		err := data.SetLog(nil)
		if err != nil {
			return
		}
	}
	if logging {
		data.Logs.Add("===== Начало расчета =====")
		defer data.Logs.Add("===== Конец расчета =====")
	}

	//Init arrays
	data.FiFo = make([]float64, len(data.Pressure)-1)
	data.Pg = make([]float64, len(data.Pressure)-1)
	data.Pn = make([]float64, len(data.Pressure)-1)
	data.Yn = make([]float64, len(data.Pressure)-1)
	data.Zn = make([]float64, len(data.Pressure)-1)
	data.Pm = make([]float64, len(data.Pressure)-1)
	data.Zm = make([]float64, len(data.Pressure)-1)
	data.Fz = make([]float64, len(data.Pressure)-1)
	data.Cn = make([]float64, len(data.Pressure)-1)
	data.G = make([]float64, len(data.Pressure)-1)
	data.Yc = make([]float64, len(data.Pressure)-1)
	data.Xi = make([]float64, len(data.Pressure)-1)
	data.Yi = make([]float64, len(data.Pressure)-1)
	data.Nf0 = make([]float64, len(data.Pressure)-1)
	data.E = make([]float64, len(data.Pressure)-1)

	data.r2 = -1
	data.se = math.Inf(1)

	// Расчет
	data.calcMu()
	data.Add("Mu", data.Mu, logging)
	data.Vp = (math.Pow(data.D, 2) * math.Pi / 4) * data.L * (data.Pori / 100) //Vp calc
	data.Add("Vp", data.Vp, logging)
	minPg := 0
	for i := 0; i < len(data.Pressure)-1; i++ {
		data.Pg[i] = math.Sqrt(data.Pressure[i] * data.Pressure[i+1]) //Pg calc
		data.Pn[i] = data.Pg[i] + data.Patm
		data.Yn[i] = data.Vt * math.Log(data.Pressure[i]/data.Pressure[i+1]) / (data.Time[i+1] - data.Time[i]) //Yn calc
		if data.Pg[i] < data.Pg[minPg] {                                                                       //Find min Pg index
			minPg = i
		}
		data.FiFo[i] = FiFoStart //Set start FiFo
	}
	firstB := 0.803 * math.Pow(math.Pow(data.D, 2)/(data.Mu*data.L*data.Yn[minPg]), 0.467) //b calc from min Pg index for Yn
	data.b = firstB
	data.Add("First b", data.b, logging)

	bestB := data.findBestB(firstB, logging)
	data.b = bestB
	mainCycle(data)
	b = data.b / 6894.76                                                               //b calc
	K = data.Mu * data.L / (math.Pi * math.Pow(data.D, 2) / 4) / data.A1 / 9.86923e-16 //K calc
	data.Logs.Add("A1=", data.A1, ";A2=", data.A2, ";SE=", data.se, ";R2=", data.r2)
	for i := range data.Xi {
		data.Logs.Add(data.Xi[i], ";", data.Yi[i])
	}
	return
}

func (data *SampleData) AssessMethod() MethodAssessment {
	assessment := MethodAssessment{
		Rows: len(data.Time),
	}

	fullSample := data.clone()
	fullK, fullB := fullSample.Calc(false)
	assessment.Full = FitResult{
		K:    fullK,
		BPsi: fullB,
		SE:   fullSample.se,
		R2:   fullSample.r2,
		A1:   fullSample.A1,
		A2:   fullSample.A2,
		BPa:  fullSample.b,
	}

	darcySample := data.clone()
	assessment.Darcy = darcySample.CalcDarcyLike(false)
	assessment.SeedBPsi = data.initialGuessBPa() / 6894.76
	assessment.MeanPmPsi = meanFloat64(darcySample.Pm) / 6894.76

	if assessment.Darcy.K > 0 {
		assessment.RelativeKGap = math.Abs(assessment.Full.K-assessment.Darcy.K) / assessment.Darcy.K
	}
	assessment.B22Valid = assessment.Darcy.R2 >= 0.88
	assessment.B22Marginal = assessment.Darcy.R2 >= 0.82 && assessment.Darcy.R2 < 0.88
	assessment.FullBCollapsed = assessment.Full.BPsi <= math.Max(0.5, assessment.SeedBPsi*0.05)
	assessment.HighPerm = math.Max(assessment.Full.K, assessment.Darcy.K) >= 100

	switch {
	case (assessment.B22Valid || assessment.B22Marginal) && assessment.FullBCollapsed && !assessment.HighPerm:
		assessment.Recommendation = "b22"
		assessment.KSource = "B-22"
		assessment.BSource = "B-22"
		assessment.RecommendedKMD = assessment.Darcy.K
		assessment.RecommendedBPsi = assessment.Darcy.B
		assessment.Rationale = []string{
			"Линейность B-22 на обработанной кривой приемлемая.",
			"Во full-fit коэффициент b схлопывается относительно стартового значения RP40.",
			"Образец не относится к диапазону высокой проницаемости, где гибридный вариант обычно полезен.",
		}
	case (assessment.B22Valid || assessment.B22Marginal) && assessment.FullBCollapsed && assessment.HighPerm && assessment.RelativeKGap >= 0.05 && assessment.RelativeKGap <= 0.20:
		assessment.Recommendation = "hybrid"
		assessment.KSource = "full"
		assessment.BSource = "B-22"
		assessment.RecommendedKMD = assessment.Full.K
		assessment.RecommendedBPsi = assessment.Darcy.B
		assessment.Rationale = []string{
			"B-22 даёт пригодную оценку скольжения, но дарси-подобная аппроксимация неидеальна для K при такой проницаемости.",
			"Full-fit сохраняет правдоподобный K, хотя b во full-fit схлопывается.",
			"Расхождение K между методами умеренное, а не критическое.",
		}
	case assessment.B22Valid && assessment.RelativeKGap <= 0.08:
		assessment.Recommendation = "b22"
		assessment.KSource = "B-22"
		assessment.BSource = "B-22"
		assessment.RecommendedKMD = assessment.Darcy.K
		assessment.RecommendedBPsi = assessment.Darcy.B
		assessment.Rationale = []string{
			"Линейность B-22 хорошая.",
			"K из full-fit и B-22 достаточно близки, поэтому предпочтительнее более простая дарси-подобная ветвь.",
		}
	case !assessment.FullBCollapsed && !assessment.B22Marginal && !assessment.B22Valid:
		assessment.Recommendation = "full"
		assessment.KSource = "full"
		assessment.BSource = "full"
		assessment.RecommendedKMD = assessment.Full.K
		assessment.RecommendedBPsi = assessment.Full.BPsi
		assessment.Rationale = []string{
			"На этой кривой линейность B-22 слабая.",
			"Во full-fit нет явного схлопывания коэффициента b.",
		}
	default:
		assessment.Recommendation = "review"
		assessment.KSource = "review"
		assessment.BSource = "review"
		if assessment.Darcy.K > 0 {
			assessment.RecommendedKMD = assessment.Darcy.K
		} else {
			assessment.RecommendedKMD = assessment.Full.K
		}
		if assessment.Darcy.B > 0 {
			assessment.RecommendedBPsi = assessment.Darcy.B
		} else {
			assessment.RecommendedBPsi = assessment.Full.BPsi
		}
		assessment.Rationale = []string{
			"Кривая попадает в неоднозначную зону между дарси-подобным и full-fit поведением.",
			"Проверьте VT, окно обработки и высоконапорную часть регрессии B-22.",
		}
	}

	return assessment
}

func (data *SampleData) CalcAtB(candidateBPa float64) FitResult {
	result := FitResult{BPa: candidateBPa, BPsi: candidateBPa / 6894.76}
	if data.Logs == nil {
		if err := data.SetLog(nil); err != nil {
			return result
		}
	}

	data.initArrays()
	data.r2 = -1
	data.se = math.Inf(1)

	data.calcMu()
	data.Vp = (math.Pow(data.D, 2) * math.Pi / 4) * data.L * (data.Pori / 100)
	for i := 0; i < len(data.Pressure)-1; i++ {
		data.Pg[i] = math.Sqrt(data.Pressure[i] * data.Pressure[i+1])
		data.Pn[i] = data.Pg[i] + data.Patm
		data.Yn[i] = data.Vt * math.Log(data.Pressure[i]/data.Pressure[i+1]) / (data.Time[i+1] - data.Time[i])
	}

	data.b = candidateBPa
	mainCycle(data)

	result.K = data.Mu * data.L / (math.Pi * math.Pow(data.D, 2) / 4) / data.A1 / 9.86923e-16
	result.SE = data.se
	result.R2 = data.r2
	result.A1 = data.A1
	result.A2 = data.A2
	return result
}

func (data *SampleData) CalcDarcyLike(logging bool) DarcyLikeResult {
	result := DarcyLikeResult{}
	if data.Logs == nil {
		if err := data.SetLog(nil); err != nil {
			return result
		}
	}

	data.initArrays()
	data.r2 = -1
	data.se = math.Inf(1)
	data.calcMu()
	data.Vp = (math.Pow(data.D, 2) * math.Pi / 4) * data.L * (data.Pori / 100)

	minPg := 0
	for i := 0; i < len(data.Pressure)-1; i++ {
		data.Pg[i] = math.Sqrt(data.Pressure[i] * data.Pressure[i+1])
		data.Pn[i] = data.Pg[i] + data.Patm
		data.Pm[i] = (data.Pg[i] / 2) + data.Patm
		data.Yn[i] = data.Vt * math.Log(data.Pressure[i]/data.Pressure[i+1]) / (data.Time[i+1] - data.Time[i])
		if data.Pg[i] < data.Pg[minPg] {
			minPg = i
		}
	}

	b := 0.803 * math.Pow(math.Pow(data.D, 2)/(data.Mu*data.L*data.Yn[minPg]), 0.467)
	slope, intercept, r2 := 0.0, 0.0, -1.0
	var finalYs []float64
	for iter := 0; iter < 20; iter++ {
		data.b = b
		data.Gamma = (2 * data.Vp) / (3 * data.Vt)
		data.calcZ()

		ys := make([]float64, len(data.Pg))
		for i := 0; i < len(data.Pg); i++ {
			data.Cn[i] = math.Pow(data.Patm+data.b, 2) / (2 * data.Pg[i] * (data.Pm[i] + data.b))
			cn := data.Cn[i]
			data.G[i] = ((cn+1)/5)*(8*cn*cn-4*cn+3) - (8 * math.Pow(cn+1, 0.5) / 5 * math.Pow(cn, 2.5))
			data.Yc[i] = data.Yn[i] * data.Fz[i] * (1 + data.Gamma*data.G[i]) / data.Zn[i]
			ys[i] = data.Yc[i] * data.Zm[i]
		}
		finalYs = append(finalYs[:0], ys...)

		slope, intercept, r2 = linearRegression(data.Pm, ys)
		if slope == 0 {
			break
		}
		newB := intercept / slope
		if newB <= 0 {
			break
		}
		if math.Abs(newB-b) <= 1e-6*math.Max(1, math.Abs(b)) {
			b = newB
			break
		}
		b = newB
	}

	if logging {
		data.Logs.Add("Darcy-like fit:", "slope=", slope, "intercept=", intercept, "b=", b, "R2=", r2)
	}
	area := math.Pi * math.Pow(data.D, 2) / 4
	result.K = data.Mu * data.L / area * slope / 9.86923e-16
	result.B = b / 6894.76
	result.R2 = r2
	result.Slope = slope
	result.Intercept = intercept
	result.Linearity = make([]B22LinearityPoint, len(data.Pm))
	for i := range data.Pm {
		fit := intercept + slope*data.Pm[i]
		residual := finalYs[i] - fit
		if math.Abs(residual) > result.MaxResidual {
			result.MaxResidual = math.Abs(residual)
		}
		result.Linearity[i] = B22LinearityPoint{
			MeanPressurePa:  data.Pm[i],
			MeanPressurePsi: data.Pm[i] / 6894.76,
			Response:        finalYs[i],
			FitResponse:     fit,
			Residual:        residual,
		}
	}
	return result
}

func (data *SampleData) initArrays() {
	size := len(data.Pressure) - 1
	data.FiFo = make([]float64, size)
	data.Pg = make([]float64, size)
	data.Pn = make([]float64, size)
	data.Yn = make([]float64, size)
	data.Zn = make([]float64, size)
	data.Pm = make([]float64, size)
	data.Zm = make([]float64, size)
	data.Fz = make([]float64, size)
	data.Cn = make([]float64, size)
	data.G = make([]float64, size)
	data.Yc = make([]float64, size)
	data.Xi = make([]float64, size)
	data.Yi = make([]float64, size)
	data.Nf0 = make([]float64, size)
	data.E = make([]float64, size)
	for i := range data.FiFo {
		data.FiFo[i] = FiFoStart
	}
}

func (data *SampleData) initialGuessBPa() float64 {
	sample := data.clone()
	sample.initArrays()
	sample.calcMu()
	minPg := 0
	for i := 0; i < len(sample.Pressure)-1; i++ {
		sample.Pg[i] = math.Sqrt(sample.Pressure[i] * sample.Pressure[i+1])
		sample.Yn[i] = sample.Vt * math.Log(sample.Pressure[i]/sample.Pressure[i+1]) / (sample.Time[i+1] - sample.Time[i])
		if sample.Pg[i] < sample.Pg[minPg] {
			minPg = i
		}
	}
	return 0.803 * math.Pow(math.Pow(sample.D, 2)/(sample.Mu*sample.L*sample.Yn[minPg]), 0.467)
}

func (data *SampleData) clone() SampleData {
	clone := *data
	clone.Time = append([]float64(nil), data.Time...)
	clone.Pressure = append([]float64(nil), data.Pressure...)
	return clone
}

func meanFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func (data *SampleData) findBestB(firstB float64, logging bool) float64 {
	bestB := firstB
	bestSE := data.evaluateB(firstB)
	if logging {
		data.Logs.Add("Initial b search point:", firstB, "SE=", bestSE)
	}

	upB, upSE := data.searchB(firstB, 1.1, bestSE, logging)
	downB, downSE := data.searchB(firstB, 0.9, bestSE, logging)

	if upSE < bestSE {
		bestB, bestSE = upB, upSE
	}
	if downSE < bestSE {
		bestB, bestSE = downB, downSE
	}

	if logging {
		data.Logs.Add("Best b selected:", bestB, "SE=", bestSE)
	}
	return bestB
}

func (data *SampleData) searchB(startB, factor, startSE float64, logging bool) (bestB float64, bestSE float64) {
	bestB, bestSE = startB, startSE
	currentB := startB

	for step := 0; step < 50; step++ {
		candidateB := currentB * factor
		if candidateB <= 0 {
			break
		}
		candidateSE := data.evaluateB(candidateB)
		if logging {
			data.Logs.Add("b search:", candidateB, "SE=", candidateSE)
		}
		if candidateSE < bestSE {
			bestB, bestSE = candidateB, candidateSE
			currentB = candidateB
			continue
		}
		break
	}

	return bestB, bestSE
}

func (data *SampleData) evaluateB(candidateB float64) float64 {
	data.b = candidateB
	mainCycle(data)
	return data.se
}
