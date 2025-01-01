package serialization

import (
	"fmt"

	"github.com/AndreiLacatos/opc-engine/node-engine/models/waveform"
	waveformvalue "github.com/AndreiLacatos/opc-engine/node-engine/models/waveform/waveform_value"
	"go.uber.org/zap"
)

type WaveformModel struct {
	Duration         int64                `json:"duration"`
	TickFrequency    int32                `json:"tickFrequency"`
	WaveformType     string               `json:"type"`
	TransitionPoints []WaveformValueModel `json:"transitionPoints"`
}

type WaveformValueModel struct {
	Tick  int64   `json:"tick"`
	Value float64 `json:"value"`
}

func (w *WaveformModel) ToDomain(l *zap.Logger) waveform.Waveform {
	waveformType := mapWaveformType(w.WaveformType, l.Named("mapper"))
	return waveform.Waveform{
		Duration:         w.Duration,
		TickFrequency:    w.TickFrequency,
		WaveformType:     waveformType,
		TransitionPoints: mapWaveformValues(w.TransitionPoints, waveformType),
	}
}

func mapWaveformType(t string, l *zap.Logger) waveform.WaveformType {
	switch t {
	case "doubleValues":
		return waveform.NumericValues
	case "transitions":
		return waveform.Transitions
	default:
		l.Warn(fmt.Sprintf("unrecognized waveform type %s, defaulting to transitions", t))
		return waveform.Transitions
	}
}

func mapWaveformValues(l []WaveformValueModel, t waveform.WaveformType) []waveform.WaveformValue {
	m := make([]waveform.WaveformValue, len(l))
	for i, v := range l {
		mappedValue := waveform.WaveformValue{
			Tick: v.Tick,
		}
		switch t {
		case waveform.Transitions:
			mappedValue.Value = &waveformvalue.Transition{}
		case waveform.NumericValues:
			mappedValue.Value = &waveformvalue.DoubleValue{Value: v.Value}
		}
		m[i] = mappedValue
	}
	return m
}
