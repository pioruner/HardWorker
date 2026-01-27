package app

import "runtime"

var (
	MacOS         bool
	MacMultiperUI float64
)

func init() {
	MacOS = runtime.GOOS == "darwin" //Check OS
	MacMultiperUI = 1.6
}
