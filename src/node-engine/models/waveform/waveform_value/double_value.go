package waveformvalue

type DoubleValue struct {
	Value float64
}

func (v *DoubleValue) GetValue() any {
	return v.Value
}
