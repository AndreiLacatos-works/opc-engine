package waveform

import waveformvalue "github.com/AndreiLacatos/opc-engine/node-engine/models/waveform/waveform_value"

type WaveformValue struct {
	Tick int64
	Value waveformvalue.WaveformPointValue
}

type WaveformType int

const (
    Transitions WaveformType = iota
    NumericValues
)

type Waveform struct {
	Duration int64
	TickFrequency int32
	WaveformType WaveformType
	TransitionPoints []WaveformValue
}
