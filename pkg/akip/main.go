package akip

func (ak *AkipW) send() {
	go ak.sendCMD()
}

func (ak *AkipW) test() {
	ak.commandInput = "STARTBIN"
	ak.send()
}

func UtoF(data []int8) []float64 {
	result := make([]float64, len(data))
	for i, v := range data {
		result[i] = float64(v)
	}
	return result
}

func BtoI(data []byte) []int8 {
	result := make([]int8, len(data))
	for i, v := range data {
		result[i] = int8(v)
	}
	return result
}
