package RP40

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pioruner/gcalc/pkg/loger"
)

func (data *SampleData) SetLog(logs *loger.Logger) error {
	if logs == nil {
		var err error
		logs, err = loger.Start("RP40")
		if err != nil {
			return err
		}
	}
	data.Logs = logs
	return nil
}

func LoadData(filename string, logs *loger.Logger, dataToLog bool) ([]float64, []float64, error) {
	logs.Add("Загрузка данных из файла")
	var time, pressure []float64

	file, err := os.Open(filename)
	logs.CheckError(err)
	defer func(file *os.File) {
		err := file.Close()
		logs.CheckError(err)
	}(file)

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		fields := strings.Split(line, "\t") // табуляция

		if len(fields) < 2 {
			logs.Add("Пропущена строка", lineNum, ": не хватает колонок")
			continue
		}

		t, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(fields[0]), ",", "."), 64)
		if logs.CheckError(err) {
			continue
		}
		p, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(fields[1]), ",", "."), 64)
		logs.CheckError(err)
		if logs.CheckError(err) {
			continue
		}
		//p *= 1000000
		time = append(time, t)
		pressure = append(pressure, p) //convert to Pa from MPa
		if dataToLog {
			logs.Add(lineNum, " : ", "Time: ", t, " Pressure: ", p)
		}
	}

	err = scanner.Err()
	if logs.CheckError(err) {
		return nil, nil, err
	}

	return time, pressure, nil
}

func (data *SampleData) Check() error {
	if data.Logs == nil {
		err := data.SetLog(nil)
		if err != nil {
			return err
		}
	}
	data.Logs.Add("Проверка и отсеивание данных")
	minDP, minDT, startFrac := data.preprocessParams()
	var time, pressure []float64
	var lastT, lastP float64
	lastT, lastP = -1, -1
	for i := 0; i < len(data.Time); i++ {
		if lastT < 0 {
			if data.Pressure[i] > Pmin && data.Pressure[i] < Pmax {
				lastP = data.Pressure[i]
				lastT = data.Time[i]
				time = append(time, data.Time[i])
				pressure = append(pressure, data.Pressure[i])
			}
		} else {
			if data.Pressure[i] < Pmin || data.Pressure[i] > Pmax || lastP <= data.Pressure[i] || lastP-data.Pressure[i] < minDP || data.Time[i]-lastT < minDT {
				continue
			}
			lastP = data.Pressure[i]
			lastT = data.Time[i]
			time = append(time, data.Time[i])
			pressure = append(pressure, data.Pressure[i])
		}
	}
	time, pressure = trimEarlyTransient(time, pressure, startFrac)
	data.Pressure = pressure
	data.Time = time
	data.Logs.Add("После проверки осталось строк - ", len(data.Time))
	if len(time) < 3 {
		return errors.New("слишком мало строк после отбора")
	}
	return nil
}

func (data *SampleData) ToSI() error {
	if data.Logs == nil {
		err := data.SetLog(nil)
		if err != nil {
			return err
		}
	}
	data.Logs.Add("Преобразование данных в формат СИ")
	//Проверка данных
	if data.D < 1 || data.L < 1 || data.Vt < 1 || data.Patm < 1 || data.Temp < 1 || data.Pori < 0 || data.Pori == 0 {
		return errors.New("некорректные данные на входе")
	}
	//data.Time=data.Time
	for i := range data.Pressure { //MPa to Pa
		data.Pressure[i] *= 1000000
	}
	data.Vt /= 1000000 //ml to l
	data.D /= 1000     //mm to m
	data.L /= 1000     //mm to m
	data.Temp += 273.1 //C to K
	data.Patm *= 1000
	return nil
}

func (data *SampleData) Add(name string, value float64, logging bool) {
	if logging {
		data.Logs.Add(name, "=", fmt.Sprintf("%e", value))
	}
}

func (data *SampleData) preprocessParams() (minDP, minDT, startFrac float64) {
	minDP = MinDP
	minDT = MinDT
	startFrac = StartFrac

	if data.LowPerm {
		minDP = MinDP / 4
		minDT = 1.0
		startFrac = 0.70
	}
	if data.MinDeltaP > 0 {
		minDP = data.MinDeltaP
	}
	if data.MinDeltaT > 0 {
		minDT = data.MinDeltaT
	}
	if data.StartAt > 0 && data.StartAt < 1 {
		startFrac = data.StartAt
	}

	return minDP, minDT, startFrac
}

func trimEarlyTransient(time, pressure []float64, startFrac float64) ([]float64, []float64) {
	if len(pressure) < 4 {
		return time, pressure
	}

	startPressure := pressure[0]
	startThreshold := startPressure * startFrac
	endThreshold := startPressure * EndFrac
	start := 0
	for start < len(pressure)-3 && pressure[start] > startThreshold {
		start++
	}

	end := len(pressure)
	for end > start+3 && pressure[end-1] < endThreshold {
		end--
	}

	return time[start:end], pressure[start:end]
}

func (data *SampleData) Print(full bool) {
	data.Logs.Add("===== Вывод данных в лог =====")
	data.Logs.Add("D = ", data.D)
	data.Logs.Add("L = ", data.L)
	data.Logs.Add("Vt = ", data.Vt)
	data.Logs.Add("FiFo = ", data.FiFo)
	data.Logs.Add("Patm = ", data.Patm)
	data.Logs.Add("Temp = ", data.Temp)
	data.Logs.Add("Pori = ", data.Pori)
	if data.GasType == He {
		data.Logs.Add("Gas type = He")
	} else if data.GasType == N {
		data.Logs.Add("Gas type = N")
	} else {
		data.Logs.Add("Gas type = Unknown")
	}

	if full {
		for i := 0; i < len(data.Time); i++ {
			data.Logs.Add(i, " - Time: ", data.Time[i], " Pressure: ", data.Pressure[i])
		}
	}
	data.Logs.Add("===== Конец вывода данных в лог =====")
}
