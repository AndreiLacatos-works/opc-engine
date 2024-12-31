package waveformvalue

type DoubleValue struct {
	Value float64
}

func (v *DoubleValue) GetValue() float64 {
	return v.Value
}