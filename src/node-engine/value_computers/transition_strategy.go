package valuecomputers

import (
	"github.com/AndreiLacatos/opc-engine/node-engine/models/waveform"
	waveformvalue "github.com/AndreiLacatos/opc-engine/node-engine/models/waveform/waveform_value"
	"go.uber.org/zap"
)

type transitionStrategyCalculator struct {
	logger         *zap.Logger
	waveform       waveform.Waveform
	stepCalculator stepSmoothingStrategyCalculator
}

func (c *transitionStrategyCalculator) Init() {
	// use the existing stepSmoothingStrategyCalculator, the trick is
	// that instead of the original waveform it is initialized with
	// a separate waveform, that is derived from the original, such
	// that, every transition has value 0 for false & 1 for true
	tp := make([]waveform.WaveformValue, len(c.waveform.TransitionPoints)+1)
	prev := int8(0)
	tp[0] = waveform.WaveformValue{
		Tick: 0,
		Value: &waveformvalue.DoubleValue{
			Value: float64(prev),
		},
	}
	prev = ^prev
	for i, p := range c.waveform.TransitionPoints {
		tp[i+1] = waveform.WaveformValue{
			Tick: p.Tick,
			Value: &waveformvalue.DoubleValue{
				Value: float64(1 & prev),
			},
		}
		prev = ^prev
	}

	derived := waveform.Waveform{
		Duration:         c.waveform.Duration,
		TickFrequency:    c.waveform.TickFrequency,
		WaveformType:     waveform.NumericValues,
		TransitionPoints: tp,
	}
	c.stepCalculator = stepSmoothingStrategyCalculator{
		logger:   c.logger,
		waveform: derived,
	}
	c.stepCalculator.Init()
}

func (c *transitionStrategyCalculator) GetValueAtTick(t int64) waveformvalue.WaveformPointValue {
	v := c.stepCalculator.GetValueAtTick(t)
	return &waveformvalue.Transition{
		Value: v.GetValue().(float64) != 0,
	}
}
