package RP40

import (
	"math"
	"sort"
)

func mainCycle(data *SampleData) {
	data.Gamma = (2 * data.Vp) / (3 * data.Vt) //Gamma calc
	data.calcZ()
	for i := 0; i < len(data.Pg); i++ {
		data.Pm[i] = (data.Pg[i] / 2) + data.Patm                                                   //Pm calc
		data.Cn[i] = math.Pow(data.Patm+data.b, 2) / (2 * data.Pg[i] * (data.Pm[i] + data.b))       //Cn calc
		Cn := data.Cn[i]                                                                            //NO NEED in a future!!! rebuild G
		data.G[i] = ((Cn+1)/5)*(8*Cn*Cn-4*Cn+3) - (8 * math.Pow(Cn+1, 0.5) / 5 * math.Pow(Cn, 2.5)) //G calc NEED REBUILD !!!
	}
	for i := 0; i < len(data.Pg); i++ {
		data.Yc[i] = data.Yn[i] * data.Fz[i] * (1 + data.Gamma*data.G[i]) / data.Zn[i] //Yc calc
		data.Xi[i] = data.Yc[i] * data.Pg[i] * (data.Pm[i] + data.b) / data.Pm[i]      //Xi calc
		data.Yi[i] = (data.Pm[i] + data.FiFo[i]*data.b) / (data.Yc[i] * data.Zm[i])    //Yi calc
	}
	data.getLine() //A1 A2 R2 find
	for {
		fifo := append([]float64(nil), data.FiFo...)
		for i := 0; i < len(data.Pg); i++ {
			data.Nf0[i] = data.Yc[i] * data.Pg[i] * data.A2 / data.A1 //Nf0 calc
			data.E[i] = data.b * data.Nf0[i] / (1 + data.Nf0[i])
			//data.FiFo[i] = ((1-data.E[i]/data.Pressure[i]*math.Log((data.Pressure[i]+data.Patm+data.E[i])/(data.Patm+data.E[i])))*((data.Pressure[i]+data.Patm*2+2*data.E[i])/(data.Pressure[i]+data.Patm*2)) + data.Nf0[i]) / (1 + data.Nf0[i])
			data.FiFo[i] = ((1-data.E[i]/data.Pg[i]*math.Log((data.Pg[i]+data.Patm+data.E[i])/(data.Patm+data.E[i])))*((data.Pg[i]+data.Patm*2+2*data.E[i])/(data.Pg[i]+data.Patm*2)) + data.Nf0[i]) / (1 + data.Nf0[i])
			data.Yi[i] = (data.Pm[i] + data.FiFo[i]*data.b) / (data.Yc[i] * data.Zm[i])
		}
		data.getLine()
		var df float64
		df = 0
		for i := 0; i < len(data.Pg); i++ {
			dt := math.Abs(fifo[i] - data.FiFo[i])
			if dt > df {
				df = dt
			}
		}
		if df < 0.001 {
			break
		}
	}

	data.Logs.Add("B=", data.b, "SE=", data.se, "R2=", data.r2)
}

func (data *SampleData) getLine() {
	n := float64(len(data.Xi))
	var sumX, sumY, sumXY, sumX2 float64
	for i := range data.Xi {
		sumX += data.Xi[i]
		sumY += data.Yi[i]
		sumXY += data.Xi[i] * data.Yi[i]
		sumX2 += data.Xi[i] * data.Xi[i]
	}
	data.A2 = (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	data.A1 = (sumY - data.A2*sumX) / n
	var errSum, mid, dv float64
	for i := range data.Xi {
		mid += data.Yi[i]
	}
	mid /= n
	for i := range data.Xi {
		yEst := data.A1 + data.A2*data.Xi[i]
		errSum += math.Pow(data.Yi[i]-yEst, 2)
		dv += math.Pow(data.Yi[i]-mid, 2)
	}
	if n > 2 {
		data.se = math.Sqrt(errSum / (n - 2))
	} else {
		data.se = math.Inf(1)
	}
	data.r2 = 1 - (errSum / dv)
}

func linearRegression(xs, ys []float64) (slope, intercept, r2 float64) {
	n := float64(len(xs))
	if len(xs) == 0 || len(xs) != len(ys) {
		return 0, 0, -1
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i := range xs {
		sumX += xs[i]
		sumY += ys[i]
		sumXY += xs[i] * ys[i]
		sumX2 += xs[i] * xs[i]
	}

	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0, 0, -1
	}

	slope = (n*sumXY - sumX*sumY) / denom
	intercept = (sumY - slope*sumX) / n

	var errSum, mid, dv float64
	for _, y := range ys {
		mid += y
	}
	mid /= n
	for i := range xs {
		yEst := intercept + slope*xs[i]
		errSum += math.Pow(ys[i]-yEst, 2)
		dv += math.Pow(ys[i]-mid, 2)
	}
	if dv == 0 {
		return slope, intercept, -1
	}
	r2 = 1 - (errSum / dv)
	return slope, intercept, r2
}

func (data *SampleData) calcMu() {
	pressureAtm := data.meanCorePressure() / 101325.0
	switch data.GasType {
	case He:
		data.Mu = heliumViscosity(data.Temp, pressureAtm)
	case N:
		data.Mu = nitrogenViscosity(data.Temp, pressureAtm)
	}
}

func (data *SampleData) meanCorePressure() float64 {
	if len(data.Pressure) < 2 {
		return data.Patm
	}

	sumPm := 0.0
	count := 0
	for i := 0; i < len(data.Pressure)-1; i++ {
		pg := math.Sqrt(data.Pressure[i] * data.Pressure[i+1])
		pm := (pg / 2) + data.Patm
		if pm <= 0 {
			continue
		}
		sumPm += pm
		count++
	}
	if count == 0 {
		return data.Patm
	}
	return sumPm / float64(count)
}

func nitrogenViscosity(tempK, pressureAtm float64) float64 {
	base := 13.85 * math.Pow(tempK, 1.5) / (tempK + 102)
	micropoise := base - 0.12474 + 0.123688*pressureAtm + 1.05452e-3*pressureAtm*pressureAtm - 1.5052e-6*pressureAtm*pressureAtm*pressureAtm
	return micropoise / 10000000
}

func heliumViscosity(tempK, pressureAtm float64) float64 {
	base := 187 * math.Pow(tempK/273.1, 0.685)
	return base * heliumViscosityCorrection(pressureAtm) / 10000000
}

func heliumViscosityCorrection(pressureAtm float64) float64 {
	switch {
	case pressureAtm <= 1:
		return 1
	case pressureAtm <= 37:
		return 1 + (pressureAtm-1)*(0.9957-1)/(37-1)
	case pressureAtm <= 158:
		return 0.9957 + (pressureAtm-37)*(1.0017-0.9957)/(158-37)
	default:
		return 1.0017
	}
}

func (data *SampleData) calcZ() {
	var points []Point
	if data.GasType == He {
		points = TableHe()
	} else {
		points = TableN()
	}

	// Собираем уникальные значения
	xVals := uniqueSorted(getField(points, func(p Point) float64 { return p.Pressure }))
	yVals := uniqueSorted(getField(points, func(p Point) float64 { return p.Temperature }))

	// Формируем сетку Zm
	zVals := make([][]float64, len(yVals))
	for i := range zVals {
		zVals[i] = make([]float64, len(xVals))
	}
	for _, p := range points {
		xi := indexExact(p.Pressure, xVals)
		yi := indexExact(p.Temperature, yVals)
		zVals[yi][xi] = p.Zm
	}

	for i := 0; i < len(data.Pm); i++ {
		data.Zm[i] = bilinearI(xVals, yVals, zVals, data.Pm[i], data.Temp, data)
		data.Zn[i] = bilinearI(xVals, yVals, zVals, data.Pn[i], data.Temp, data)
		data.Fz[i] = calcFz(xVals, yVals, zVals, data.Pn[i], data.Temp, data, data.Zn[i])
	}
}

func calcFz(xVals, yVals []float64, zVals [][]float64, pressure, temperature float64, data *SampleData, z float64) float64 {
	if z == 0 {
		return 1
	}

	i := indexRange(pressure, xVals)
	j := indexRange(temperature, yVals)
	if i == -1 || j == -1 {
		return 1
	}

	x1, x2 := xVals[i], xVals[i+1]
	if x2 == x1 {
		return 1
	}

	y1, y2 := yVals[j], yVals[j+1]
	ty := 0.0
	if y2 != y1 {
		ty = (temperature - y1) / (y2 - y1)
	}

	lowSlope := (zVals[j][i+1] - zVals[j][i]) / (x2 - x1)
	highSlope := (zVals[j+1][i+1] - zVals[j+1][i]) / (x2 - x1)
	dzdp := lowSlope*(1-ty) + highSlope*ty
	fz := 1 - (pressure/z)*dzdp
	if fz <= 0 {
		data.Logs.Add("fz became non-positive, fallback to 1:", fz, "at P=", pressure)
		return 1
	}
	return fz
}

func bilinearI(xVals, yVals []float64, zVals [][]float64, x, y float64, data *SampleData) float64 {
	i := indexRange(x, xVals)
	j := indexRange(y, yVals)

	if i == -1 || j == -1 {
		data.Logs.AddF("Interpolation out of bounds: x=", x, " y=", y)
	}

	x1, x2 := xVals[i], xVals[i+1]
	y1, y2 := yVals[j], yVals[j+1]

	z11 := zVals[j][i]
	z21 := zVals[j][i+1]
	z12 := zVals[j+1][i]
	z22 := zVals[j+1][i+1]

	tx := (x - x1) / (x2 - x1)
	ty := (y - y1) / (y2 - y1)

	return z11*(1-tx)*(1-ty) + z21*tx*(1-ty) + z12*(1-tx)*ty + z22*tx*ty
}

// Уникальные отсортированные значения
func uniqueSorted(data []float64) []float64 {
	m := map[float64]bool{}
	for _, v := range data {
		m[v] = true
	}
	var res []float64
	for v := range m {
		res = append(res, v)
	}
	sort.Float64s(res)
	return res
}

func getField(data []Point, f func(Point) float64) []float64 {
	res := make([]float64, len(data))
	for i, p := range data {
		res[i] = f(p)
	}
	return res
}

func indexExact(val float64, arr []float64) int {
	for i, v := range arr {
		if v == val {
			return i
		}
	}
	return -1
}

func indexRange(val float64, arr []float64) int {
	for i := 0; i < len(arr)-1; i++ {
		if arr[i] <= val && val <= arr[i+1] {
			return i
		}
	}
	return -1
}
