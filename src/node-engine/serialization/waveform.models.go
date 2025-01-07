package serialization

import (
	"fmt"
	"strings"

	"github.com/AndreiLacatos/opc-engine/node-engine/models/waveform"
	waveformvalue "github.com/AndreiLacatos/opc-engine/node-engine/models/waveform/waveform_value"
	"go.uber.org/zap"
)

type WaveformModel struct {
	Duration         int64                `json:"duration"`
	TickFrequency    int32                `json:"tickFrequency"`
	WaveformType     string               `json:"type"`
	TransitionPoints []WaveformValueModel `json:"transitionPoints"`
	Meta             *WaveformMetaModel   `json:"meta"`
}

type WaveformValueModel struct {
	Tick  int64   `json:"tick"`
	Value float64 `json:"value"`
}

type WaveformMetaModel struct {
	Smoothing *string `json:"smoothing"`
}

func (w *WaveformModel) ToDomain(l *zap.Logger) waveform.Waveform {
	waveformType := mapWaveformType(w.WaveformType, l.Named("mapper"))
	return waveform.Waveform{
		Duration:         w.Duration,
		TickFrequency:    w.TickFrequency,
		WaveformType:     waveformType,
		TransitionPoints: mapWaveformValues(w.TransitionPoints, waveformType),
		Meta:             mapWaveformMeta(w.Meta, waveformType, l),
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

func mapWaveformMeta(m *WaveformMetaModel, t waveform.WaveformType, l *zap.Logger) *waveform.WaveformMeta {

	switch t {
	case waveform.Transitions:
		return nil
	case waveform.NumericValues:
		if m == nil {
			l.Warn(fmt.Sprintf("missing meta, using defaults for type %v", t))
			var d waveform.WaveformMeta = waveform.NumericWaveformMeta{
				Smoothing: waveform.Step,
			}
			return &d
		}
		if m.Smoothing == nil {
			l.Warn("missing smoothing type, using default")
			return nil
		}
		var s waveform.SmoothingStrategy
		switch strings.ToLower(*m.Smoothing) {
		case "step":
			s = waveform.Step
		case "linear":
			s = waveform.Linear
		case "cubic":
			s = waveform.CubicSpline
		}
		var m waveform.WaveformMeta = waveform.NumericWaveformMeta{
			Smoothing: s,
		}
		return &m
	}
	return nil
}
