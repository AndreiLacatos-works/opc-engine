package delaycalculator

import (
	"time"

	"github.com/AndreiLacatos/opc-engine/node-engine/models/waveform"
)

type delayCalculatorImpl struct {
	waveform      waveform.Waveform
	nextTickIndex int
	tickSchedule  []time.Time
}

func (c *delayCalculatorImpl) init(w waveform.Waveform) {
	c.waveform = w
	c.nextTickIndex = 0
}

func (c *delayCalculatorImpl) GetDelayUntilNextTick() time.Duration {
	if c.tickSchedule == nil || c.nextTickIndex >= len(c.tickSchedule) {
		c.makeCycleSchedule()
		c.nextTickIndex = 0
	}

	delay := time.Until(c.tickSchedule[c.nextTickIndex])
	c.nextTickIndex += 1
	return delay
}

func (c *delayCalculatorImpl) makeCycleSchedule() {
	tickCount := c.waveform.Duration / int64(c.waveform.TickFrequency)
	scheduleLength := tickCount
	if c.waveform.Duration%int64(c.waveform.TickFrequency) != 0 {
		scheduleLength += 1
	}

	schedule := make([]time.Time, 0, scheduleLength)
	startTime := time.Now()

	// schedule ticks for the next 10 cycles
	for j := 0; j < 10; j++ {
		// schedule tick for an entire cycle
		for i := int64(1); i <= tickCount; i++ {
			tickDelay := time.Duration(i*int64(c.waveform.TickFrequency)) * time.Millisecond
			schedule = append(schedule, startTime.Add(tickDelay))
		}
		if c.waveform.Duration%int64(c.waveform.TickFrequency) != 0 {
			// add an extra scheduled date when the last tick does not intersect
			// with the duration of the cycle
			cycleEnd := time.Duration(c.waveform.Duration) * time.Millisecond
			schedule = append(schedule, startTime.Add(cycleEnd))
		}

		// compute the start time of the next cycle
		startTime = startTime.Add(time.Duration(c.waveform.Duration) * time.Millisecond)
	}

	c.tickSchedule = schedule
}
