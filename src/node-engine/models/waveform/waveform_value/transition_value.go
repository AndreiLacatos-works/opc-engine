package waveformvalue

type Transition struct {
	Value bool
}

func (t *Transition) GetValue() any {
	return t.Value
}
