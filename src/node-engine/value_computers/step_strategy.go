package valuecomputers

import (
	"fmt"

	"github.com/AndreiLacatos/opc-engine/node-engine/models/waveform"
	waveformvalue "github.com/AndreiLacatos/opc-engine/node-engine/models/waveform/waveform_value"
	"go.uber.org/zap"
)

type stepSmoothingStrategyCalculator struct {
	logger   *zap.Logger
	waveform waveform.Waveform
	values   map[int64]waveformvalue.WaveformPointValue
}

func (c *stepSmoothingStrategyCalculator) Init() {
	c.values = make(map[int64]waveformvalue.WaveformPointValue)
	// add the explicit transitions
	for _, t := range c.waveform.TransitionPoints {
		c.values[t.Tick] = t.Value
	}

	// add artificial entries for the ticks that do not have explicit transitions
	prev := c.waveform.TransitionPoints[0].Value
	tickCount := c.waveform.Duration / int64(c.waveform.TickFrequency)
	for i := int64(0); i <= tickCount; i++ {
		t := i * int64(c.waveform.TickFrequency)
		if v, found := c.values[t]; found {
			prev = v
		} else {
			c.values[t] = prev
		}
	}
}

func (c *stepSmoothingStrategyCalculator) GetValueAtTick(t int64) waveformvalue.WaveformPointValue {
	if v, found := c.values[t]; !found {
		c.logger.Warn(fmt.Sprintf("invalid tick %d", t))
		return &waveformvalue.DoubleValue{Value: 0.0}
	} else {
		return v
	}
}
