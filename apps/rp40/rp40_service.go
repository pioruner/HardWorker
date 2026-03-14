package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	calc "github.com/pioruner/gcalc/pkg/RP40"
	"golang.org/x/text/encoding/charmap"
)

type LogEntry struct {
	Time    string
	Level   string
	Message string
}

type PassportSampleOption struct {
	ID          string
	Label       string
	Occurrences int
	UpdatedAt   string
}

type PassportSampleRecord struct {
	SampleID          string
	Timestamp         string
	Gas               string
	LengthMM          float64
	DiameterMM        float64
	HeightMM          float64
	VolumeCM3         float64
	PorosityPct       float64
	TemperatureC      float64
	AtmosphericKPa    float64
	CompressionMPa    float64
	PoreVolumeCM3     float64
	KGasMD            float64
	KKlinkenbergMD    float64
	BPsi              float64
	RawRow            int
	SourceOccurrences int
}

type RP40Inputs struct {
	SampleID          string
	Gas               string
	DiameterMM        float64
	LengthMM          float64
	PorosityPct       float64
	TemperatureC      float64
	AtmosphericKPa    float64
	ReservoirMode     string
	ReservoirVolumeML float64
	CustomReservoirML float64
	LowPerm           bool
	MinDeltaP         float64
	MinDeltaT         float64
	StartAt           float64
}

type ChartPoint struct {
	X float64
	Y float64
}

type B22LinearityPoint struct {
	MeanPressurePa  float64
	MeanPressurePsi float64
	Response        float64
	FitResponse     float64
	Residual        float64
}

type FullFitView struct {
	KMD  float64
	BPa  float64
	BPsi float64
	SE   float64
	R2   float64
	A1   float64
	A2   float64
}

type B22View struct {
	KMD         float64
	BPsi        float64
	R2          float64
	Slope       float64
	Intercept   float64
	MaxResidual float64
	Linearity   []B22LinearityPoint
}

type AssessmentView struct {
	Recommendation  string
	KSource         string
	BSource         string
	RecommendedKMD  float64
	RecommendedBPsi float64
	Rationale       []string
	B22Valid        bool
	B22Marginal     bool
	FullBCollapsed  bool
	HighPerm        bool
	RelativeKGap    float64
	MeanPmPsi       float64
}

type CalculationView struct {
	HasResult       bool
	ProcessedRows   int
	DisplayKMD      float64
	DisplayBPsi     float64
	DisplayMethod   string
	DisplayWarnings []string
	ProcessedCurve  []ChartPoint
	B22FitCurve     []ChartPoint
	Full            FullFitView
	B22             B22View
	Assessment      AssessmentView
}

type RP40Snapshot struct {
	PassportFilePath    string
	MeasurementFilePath string
	Samples             []PassportSampleOption
	SelectedSample      PassportSampleRecord
	Inputs              RP40Inputs
	Calculation         CalculationView
	Logs                []LogEntry
	LastError           string
}

type passportRow struct {
	record      PassportSampleRecord
	parsedStamp time.Time
}

type RP40Service struct {
	mu              sync.RWMutex
	ctx             context.Context
	passportPath    string
	measurementPath string
	samples         []PassportSampleOption
	records         map[string]PassportSampleRecord
	selectedID      string
	inputs          RP40Inputs
	calculation     CalculationView
	logs            []LogEntry
	lastError       string
}

var reservoirOptions = map[string]float64{
	"17.3": 17.3,
	"233":  233,
	"1445": 1445,
}

func NewRP40Service() *RP40Service {
	svc := &RP40Service{
		records: make(map[string]PassportSampleRecord),
		inputs: RP40Inputs{
			ReservoirMode:     "233",
			ReservoirVolumeML: 233,
			StartAt:           0.85,
		},
	}
	svc.log("INFO", "RP40 service initialized")
	return svc
}

func (s *RP40Service) SetContext(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ctx = ctx
}

func (s *RP40Service) GetSnapshot() RP40Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	samples := append([]PassportSampleOption(nil), s.samples...)
	logs := append([]LogEntry(nil), s.logs...)
	return RP40Snapshot{
		PassportFilePath:    s.passportPath,
		MeasurementFilePath: s.measurementPath,
		Samples:             samples,
		SelectedSample:      s.records[s.selectedID],
		Inputs:              s.inputs,
		Calculation:         cloneCalculation(s.calculation),
		Logs:                logs,
		LastError:           s.lastError,
	}
}

func (s *RP40Service) LoadPassport(path string) {
	rows, err := parsePassport(path)
	s.mu.Lock()
	defer s.mu.Unlock()

	if err != nil {
		s.lastError = err.Error()
		s.log("ERROR", "Passport load failed: "+err.Error())
		return
	}

	s.passportPath = path
	s.records = make(map[string]PassportSampleRecord)
	s.samples = s.samples[:0]
	s.selectedID = ""
	s.lastError = ""

	for _, row := range rows {
		s.records[row.record.SampleID] = row.record
		s.samples = append(s.samples, PassportSampleOption{
			ID:          row.record.SampleID,
			Label:       row.record.SampleID,
			Occurrences: row.record.SourceOccurrences,
			UpdatedAt:   row.record.Timestamp,
		})
	}

	if len(s.samples) > 0 {
		s.selectedID = s.samples[0].ID
		s.applyRecordLocked(s.records[s.selectedID])
	}

	s.log("INFO", fmt.Sprintf("Passport loaded: %s (%d samples)", filepath.Base(path), len(s.samples)))
}

func (s *RP40Service) SetMeasurementFile(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.measurementPath = path
	s.lastError = ""
	if matchedID, ok := s.matchSampleByMeasurementPathLocked(path); ok {
		s.selectedID = matchedID
		s.applyRecordLocked(s.records[matchedID])
		s.log("INFO", "Auto-selected sample from measurement file: "+matchedID)
	}
	s.log("INFO", "Measurement file selected: "+filepath.Base(path))
}

func (s *RP40Service) SetSelectedSample(sampleID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[sampleID]
	if !ok {
		s.lastError = "sample not found in loaded passport"
		s.log("WARN", "Attempt to select unknown sample: "+sampleID)
		return
	}
	s.selectedID = sampleID
	s.applyRecordLocked(record)
	s.lastError = ""
	s.log("INFO", "Selected sample: "+sampleID)
}

func (s *RP40Service) UpdateInputs(in RP40Inputs) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.inputs = in
	s.inputs.SampleID = s.selectedID
	s.inputs.ReservoirVolumeML = resolveReservoirVolume(in)
	s.lastError = ""
}

func (s *RP40Service) Calculate() {
	s.mu.Lock()
	inputs := s.inputs
	inputs.SampleID = s.selectedID
	inputs.ReservoirVolumeML = resolveReservoirVolume(inputs)
	measurementPath := s.measurementPath
	s.inputs = inputs
	s.lastError = ""
	s.mu.Unlock()

	result, err := buildCalculation(inputs, measurementPath)

	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		s.lastError = err.Error()
		s.calculation = CalculationView{}
		s.log("ERROR", "Calculation failed: "+err.Error())
		return
	}

	s.calculation = result
	s.lastError = ""
	s.log("INFO", fmt.Sprintf("Calculation finished: method=%s K=%.4f b=%.4f", result.DisplayMethod, result.DisplayKMD, result.DisplayBPsi))
}

func (s *RP40Service) applyRecordLocked(record PassportSampleRecord) {
	s.inputs.SampleID = record.SampleID
	s.inputs.Gas = record.Gas
	s.inputs.DiameterMM = record.DiameterMM
	s.inputs.LengthMM = record.LengthMM
	s.inputs.PorosityPct = record.PorosityPct
	s.inputs.TemperatureC = record.TemperatureC
	s.inputs.AtmosphericKPa = record.AtmosphericKPa
	if s.inputs.ReservoirMode == "" {
		s.inputs.ReservoirMode = "233"
	}
	s.inputs.ReservoirVolumeML = resolveReservoirVolume(s.inputs)
}

func (s *RP40Service) log(level, message string) {
	entry := LogEntry{
		Time:    time.Now().Format("15:04:05"),
		Level:   level,
		Message: message,
	}
	s.logs = append(s.logs, entry)
	if len(s.logs) > 250 {
		s.logs = append([]LogEntry(nil), s.logs[len(s.logs)-250:]...)
	}
}

func (s *RP40Service) matchSampleByMeasurementPathLocked(path string) (string, bool) {
	if len(s.records) == 0 {
		return "", false
	}

	baseName := strings.TrimSpace(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
	if baseName == "" {
		return "", false
	}

	candidates := []string{
		baseName,
		strings.TrimSpace(strings.Split(baseName, " - ")[0]),
		strings.TrimSpace(strings.Split(baseName, "_")[0]),
	}

	normalizedCandidates := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		normalized := normalizeSampleKey(candidate)
		if normalized != "" {
			normalizedCandidates = append(normalizedCandidates, normalized)
		}
	}

	bestID := ""
	bestScore := 0
	for sampleID := range s.records {
		normalizedSample := normalizeSampleKey(sampleID)
		if normalizedSample == "" {
			continue
		}
		for _, candidate := range normalizedCandidates {
			score := sampleMatchScore(normalizedSample, candidate)
			if score > bestScore {
				bestScore = score
				bestID = sampleID
			}
		}
	}

	return bestID, bestScore > 0
}

func buildCalculation(inputs RP40Inputs, measurementPath string) (CalculationView, error) {
	if strings.TrimSpace(measurementPath) == "" {
		return CalculationView{}, errors.New("не загружен файл замера")
	}
	if strings.TrimSpace(inputs.SampleID) == "" {
		return CalculationView{}, errors.New("не выбран образец из паспортного файла")
	}
	if inputs.DiameterMM <= 0 || inputs.LengthMM <= 0 || inputs.PorosityPct <= 0 || inputs.TemperatureC <= 0 || inputs.AtmosphericKPa <= 0 {
		return CalculationView{}, errors.New("проверьте размеры, пористость, температуру и Pатм")
	}

	reservoir := resolveReservoirVolume(inputs)
	if reservoir <= 0 {
		return CalculationView{}, errors.New("некорректный объем Vt")
	}

	analyzeResult, err := calc.Analyze(calc.AnalyzeRequest{
		DataFile:          measurementPath,
		DiameterMM:        inputs.DiameterMM,
		LengthMM:          inputs.LengthMM,
		ReservoirVolumeML: reservoir,
		AtmosphericKPa:    inputs.AtmosphericKPa,
		TemperatureC:      inputs.TemperatureC,
		PorosityPct:       inputs.PorosityPct,
		Gas:               inputs.Gas,
		LowPerm:           inputs.LowPerm,
		MinDeltaP:         inputs.MinDeltaP,
		MinDeltaT:         inputs.MinDeltaT,
		StartAt:           inputs.StartAt,
	})
	if err != nil {
		return CalculationView{}, err
	}

	view := CalculationView{
		HasResult:       true,
		ProcessedRows:   analyzeResult.ProcessedRows,
		DisplayKMD:      analyzeResult.DisplayKMD,
		DisplayBPsi:     analyzeResult.DisplayBPsi,
		DisplayMethod:   analyzeResult.DisplayMethod,
		DisplayWarnings: append([]string(nil), analyzeResult.DisplayWarnings...),
		ProcessedCurve:  make([]ChartPoint, 0, len(analyzeResult.ProcessedTimeSec)),
		B22FitCurve:     make([]ChartPoint, 0, len(analyzeResult.B22.Linearity)),
		Full: FullFitView{
			KMD:  analyzeResult.Full.K,
			BPa:  analyzeResult.Full.BPa,
			BPsi: analyzeResult.Full.BPsi,
			SE:   analyzeResult.Full.SE,
			R2:   analyzeResult.Full.R2,
			A1:   analyzeResult.Full.A1,
			A2:   analyzeResult.Full.A2,
		},
		B22: B22View{
			KMD:         analyzeResult.B22.K,
			BPsi:        analyzeResult.B22.B,
			R2:          analyzeResult.B22.R2,
			Slope:       analyzeResult.B22.Slope,
			Intercept:   analyzeResult.B22.Intercept,
			MaxResidual: analyzeResult.B22.MaxResidual,
			Linearity:   make([]B22LinearityPoint, 0, len(analyzeResult.B22.Linearity)),
		},
		Assessment: AssessmentView{
			Recommendation:  analyzeResult.Assessment.Recommendation,
			KSource:         analyzeResult.Assessment.KSource,
			BSource:         analyzeResult.Assessment.BSource,
			RecommendedKMD:  analyzeResult.Assessment.RecommendedKMD,
			RecommendedBPsi: analyzeResult.Assessment.RecommendedBPsi,
			Rationale:       append([]string(nil), analyzeResult.Assessment.Rationale...),
			B22Valid:        analyzeResult.Assessment.B22Valid,
			B22Marginal:     analyzeResult.Assessment.B22Marginal,
			FullBCollapsed:  analyzeResult.Assessment.FullBCollapsed,
			HighPerm:        analyzeResult.Assessment.HighPerm,
			RelativeKGap:    analyzeResult.Assessment.RelativeKGap,
			MeanPmPsi:       analyzeResult.Assessment.MeanPmPsi,
		},
	}

	for i := range analyzeResult.ProcessedTimeSec {
		view.ProcessedCurve = append(view.ProcessedCurve, ChartPoint{
			X: analyzeResult.ProcessedTimeSec[i],
			Y: analyzeResult.ProcessedPressureM[i],
		})
	}
	for _, point := range analyzeResult.B22.Linearity {
		view.B22.Linearity = append(view.B22.Linearity, B22LinearityPoint{
			MeanPressurePa:  point.MeanPressurePa,
			MeanPressurePsi: point.MeanPressurePsi,
			Response:        point.Response,
			FitResponse:     point.FitResponse,
			Residual:        point.Residual,
		})
		view.B22FitCurve = append(view.B22FitCurve, ChartPoint{
			X: point.MeanPressurePsi,
			Y: point.FitResponse,
		})
	}

	sort.Slice(view.B22FitCurve, func(i, j int) bool {
		return view.B22FitCurve[i].X < view.B22FitCurve[j].X
	})

	return view, nil
}

func parsePassport(path string) ([]passportRow, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoded, err := charmap.Windows1251.NewDecoder().Bytes(body)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.ReplaceAll(string(decoded), "\r\n", "\n"), "\n")
	headerIndex := -1
	var headers []string
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		headers = splitTSV(line)
		headerIndex = i
		break
	}
	if headerIndex < 0 {
		return nil, errors.New("паспортный файл пустой")
	}

	columns, err := resolvePassportColumns(headers)
	if err != nil {
		return nil, err
	}

	latest := make(map[string]passportRow)
	counts := make(map[string]int)
	order := make([]string, 0)

	for lineNo := headerIndex + 1; lineNo < len(lines); lineNo++ {
		line := strings.TrimSpace(lines[lineNo])
		if line == "" {
			continue
		}
		fields := splitTSV(lines[lineNo])
		record := PassportSampleRecord{
			SampleID:       fieldValue(fields, columns["sample"]),
			Timestamp:      fieldValue(fields, columns["timestamp"]),
			Gas:            fieldValue(fields, columns["gas"]),
			LengthMM:       parseFloatField(fieldValue(fields, columns["length"])),
			DiameterMM:     parseFloatField(fieldValue(fields, columns["diameter"])),
			HeightMM:       parseFloatField(fieldValue(fields, columnIndex(columns, "height"))),
			VolumeCM3:      parseFloatField(fieldValue(fields, columnIndex(columns, "volume"))),
			PorosityPct:    parseFloatField(fieldValue(fields, columns["porosity"])),
			TemperatureC:   parseFloatField(fieldValue(fields, columns["temperature"])),
			AtmosphericKPa: parseFloatField(fieldValue(fields, columns["patm"])),
			CompressionMPa: parseFloatField(fieldValue(fields, columnIndex(columns, "compression"))),
			PoreVolumeCM3:  parseFloatField(fieldValue(fields, columnIndex(columns, "pore_volume"))),
			KGasMD:         parseFloatField(fieldValue(fields, columnIndex(columns, "kgas"))),
			KKlinkenbergMD: parseFloatField(fieldValue(fields, columnIndex(columns, "kklink"))),
			BPsi:           parseFloatField(fieldValue(fields, columnIndex(columns, "bpsi"))),
			RawRow:         lineNo + 1,
		}
		record.SampleID = strings.TrimSpace(record.SampleID)
		if record.SampleID == "" {
			continue
		}
		counts[record.SampleID]++
		if _, exists := latest[record.SampleID]; !exists {
			order = append(order, record.SampleID)
		}

		row := passportRow{
			record:      record,
			parsedStamp: parsePassportTime(record.Timestamp),
		}

		current, exists := latest[record.SampleID]
		if !exists || row.parsedStamp.After(current.parsedStamp) || row.parsedStamp.IsZero() {
			latest[record.SampleID] = row
		}
	}

	rows := make([]passportRow, 0, len(order))
	for _, id := range order {
		row := latest[id]
		row.record.SourceOccurrences = counts[id]
		rows = append(rows, row)
	}

	return rows, nil
}

func resolvePassportColumns(headers []string) (map[string]int, error) {
	normalized := make(map[string]int, len(headers))
	for idx, raw := range headers {
		normalized[normalizeHeader(raw)] = idx
	}

	columns := make(map[string]int)

	requiredSpecs := []struct {
		key      string
		aliases  []string
		fallback int
	}{
		{key: "timestamp", aliases: []string{"дата/время", "датавремя"}, fallback: 0},
		{key: "sample", aliases: []string{"№образца", "образец", "номеробразца"}, fallback: 1},
		{key: "gas", aliases: []string{"газ"}, fallback: 2},
		{key: "length", aliases: []string{"длина,мм", "длинамм"}, fallback: 3},
		{key: "diameter", aliases: []string{"диаметр(ширина),мм", "диаметрширина,мм", "диаметрмм"}, fallback: 4},
		{key: "porosity", aliases: []string{"пористость,%", "пористость%"}, fallback: 10},
		{key: "temperature", aliases: []string{"текущяят,°с", "текущяят,°c", "текущяят,с", "текущаят,°с", "текущаят,°c"}, fallback: 15},
		{key: "patm", aliases: []string{"pатм,кпа", "патм,кпа", "pатмкпа"}, fallback: 16},
	}

	optionalSpecs := []struct {
		key      string
		aliases  []string
		fallback int
	}{
		{key: "height", aliases: []string{"высота,мм", "высотамм"}, fallback: 5},
		{key: "volume", aliases: []string{"объемобразца,см^3", "объемобразца,см3"}, fallback: 6},
		{key: "compression", aliases: []string{"робжима,мпа", "робжима,мрa"}, fallback: 8},
		{key: "pore_volume", aliases: []string{"объемпор,см3", "объемпор,см^3"}, fallback: 9},
		{key: "kgas", aliases: []string{"кгаз,мд", "kgas,md"}, fallback: 12},
		{key: "kklink", aliases: []string{"ккл,мд", "kkл,мд"}, fallback: 13},
		{key: "bpsi", aliases: []string{"b,psi", "b,psi."}, fallback: 14},
	}

	for _, spec := range requiredSpecs {
		index, ok := findColumnIndex(normalized, headers, spec.aliases, spec.fallback)
		if !ok {
			return nil, fmt.Errorf("в паспортном файле не найдена обязательная колонка %q", headersHint(spec.aliases))
		}
		columns[spec.key] = index
	}

	for _, spec := range optionalSpecs {
		if index, ok := findColumnIndex(normalized, headers, spec.aliases, spec.fallback); ok {
			columns[spec.key] = index
		}
	}

	return columns, nil
}

func findColumnIndex(normalized map[string]int, headers []string, aliases []string, fallback int) (int, bool) {
	for _, alias := range aliases {
		if index, ok := normalized[normalizeHeader(alias)]; ok {
			return index, true
		}
	}
	if fallback >= 0 && fallback < len(headers) {
		return fallback, true
	}
	return -1, false
}

func headersHint(aliases []string) string {
	if len(aliases) == 0 {
		return ""
	}
	return aliases[0]
}

func resolveReservoirVolume(in RP40Inputs) float64 {
	mode := strings.TrimSpace(in.ReservoirMode)
	if mode == "custom" {
		return in.CustomReservoirML
	}
	if value, ok := reservoirOptions[mode]; ok {
		return value
	}
	if in.ReservoirVolumeML > 0 {
		return in.ReservoirVolumeML
	}
	return 0
}

func splitTSV(line string) []string {
	raw := strings.Split(strings.TrimPrefix(line, "\ufeff"), "\t")
	out := make([]string, len(raw))
	for i, item := range raw {
		out[i] = strings.TrimSpace(item)
	}
	return out
}

func fieldValue(fields []string, index int) string {
	if index < 0 || index >= len(fields) {
		return ""
	}
	return strings.TrimSpace(fields[index])
}

func columnIndex(columns map[string]int, key string) int {
	value, ok := columns[key]
	if !ok {
		return -1
	}
	return value
}

func parseFloatField(raw string) float64 {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return 0
	}
	clean = strings.ReplaceAll(clean, " ", "")
	clean = strings.ReplaceAll(clean, ",", ".")
	value, err := strconv.ParseFloat(clean, 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return value
}

func normalizeHeader(raw string) string {
	clean := strings.ToLower(strings.TrimSpace(raw))
	replacer := strings.NewReplacer(
		" ", "",
		"\u00a0", "",
		"\t", "",
		"\ufeff", "",
		"(", "",
		")", "",
		".", "",
	)
	return replacer.Replace(clean)
}

func normalizeSampleKey(raw string) string {
	clean := strings.ToLower(strings.TrimSpace(raw))
	replacer := strings.NewReplacer(
		" ", "",
		"\u00a0", "",
		"\t", "",
		"\ufeff", "",
		"_", "",
		"-", "",
		"№", "",
	)
	return replacer.Replace(clean)
}

func sampleMatchScore(sample, candidate string) int {
	if sample == "" || candidate == "" {
		return 0
	}
	if sample == candidate {
		return 1000 + len(sample)
	}
	if strings.HasPrefix(candidate, sample) {
		return 800 + len(sample)
	}
	if strings.HasPrefix(sample, candidate) {
		return 700 + len(candidate)
	}
	if strings.Contains(candidate, sample) {
		return 500 + len(sample)
	}
	if strings.Contains(sample, candidate) {
		return 400 + len(candidate)
	}
	return 0
}

func parsePassportTime(raw string) time.Time {
	normalized := strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
	layouts := []string{
		"02.01.2006 15:04:05",
		"2.1.2006 15:04:05",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, normalized); err == nil {
			return ts
		}
	}
	return time.Time{}
}

func cloneCalculation(in CalculationView) CalculationView {
	out := in
	out.DisplayWarnings = append([]string(nil), in.DisplayWarnings...)
	out.ProcessedCurve = append([]ChartPoint(nil), in.ProcessedCurve...)
	out.B22FitCurve = append([]ChartPoint(nil), in.B22FitCurve...)
	out.B22.Linearity = append([]B22LinearityPoint(nil), in.B22.Linearity...)
	out.Assessment.Rationale = append([]string(nil), in.Assessment.Rationale...)
	return out
}
