package delaycalculator

import (
	"time"

	"github.com/AndreiLacatos/opc-engine/node-engine/models/waveform"
)

type DelayCalculator interface {
	init(waveform.Waveform)
	GetDelayUntilNextTick() time.Duration
}

func CreateNew(w waveform.Waveform) DelayCalculator {
	c := delayCalculatorImpl{}
	c.init(w)
	return &c
}
