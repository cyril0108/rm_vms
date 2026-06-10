package utils

import (
	"strconv"
)

type fftime struct {
	MS int64
	FloatTime float64
}

func NewFFTime(time int64) *fftime {
	f := fftime {
		MS: time,
	}
	f.ToFloatTime()
	return &f
}

func (f *fftime) ToFloatTime() {
	f.FloatTime = float64(f.MS) / 1000.0
}

// print "3.111" means 3111 milliseconds
func (f *fftime) TimeString() string {
	return strconv.FormatFloat(f.FloatTime, 'f', 3, 64)
}