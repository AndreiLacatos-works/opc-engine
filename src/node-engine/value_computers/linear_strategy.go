package valuecomputers

import (
	"fmt"
	"sort"

	"github.com/AndreiLacatos/opc-engine/node-engine/models/waveform"
	waveformvalue "github.com/AndreiLacatos/opc-engine/node-engine/models/waveform/waveform_value"
	"go.uber.org/zap"
)

type section struct {
	from waveform.WaveformValue
	to   waveform.WaveformValue
}

// sort.Interface implementations for sorting by from.Tick
type SortedByFrom []section

func (a SortedByFrom) Len() int           { return len(a) }
func (a SortedByFrom) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortedByFrom) Less(i, j int) bool { return a[i].from.Tick < a[j].from.Tick }

type linearSmoothingStrategyCalculator struct {
	logger   *zap.Logger
	waveform waveform.Waveform
	sections []section
}

func (c *linearSmoothingStrategyCalculator) Init() {
	c.sections = make([]section, 0)
	// construct sections, where each from & to are
	// explicit transition points of the waveform
	if len(c.waveform.TransitionPoints) < 2 {
		return
	}

	prev := c.waveform.TransitionPoints[0]
	for i := 1; i < len(c.waveform.TransitionPoints); i++ {
		cur := c.waveform.TransitionPoints[i]
		c.sections = append(c.sections, section{
			from: prev,
			to:   cur,
		})
		prev = cur
	}

	first := c.waveform.TransitionPoints[0]
	if first.Tick > 0 {
		// add extra section from 0 to the first explicit transition
		c.sections = append(c.sections, section{
			from: waveform.WaveformValue{
				Tick:  0,
				Value: first.Value,
			},
			to: first,
		})
	}

	last := c.waveform.TransitionPoints[len(c.waveform.TransitionPoints)-1]
	if last.Tick < c.waveform.Duration {
		// add extra section from the last explicit transition to
		// the end of the waveform
		c.sections = append(c.sections, section{
			from: last,
			to: waveform.WaveformValue{
				Tick:  c.waveform.Duration,
				Value: last.Value,
			},
		})
	}

	// sort sections in ascending order by from.Tick
	sort.Sort(SortedByFrom(c.sections))
}

func (c *linearSmoothingStrategyCalculator) GetValueAtTick(t int64) waveformvalue.WaveformPointValue {
	s := c.GetEncompassingSection(t)
	if s == nil {
		c.logger.Warn(fmt.Sprintf("invalid tick %d", t))
		return &waveformvalue.DoubleValue{Value: 0.0}
	}
	v := mapValueToNewRange(float64(s.from.Tick), float64(s.to.Tick), float64(t),
		s.from.Value.GetValue().(float64), s.to.Value.GetValue().(float64))

	return &waveformvalue.DoubleValue{Value: v}
}

func (c linearSmoothingStrategyCalculator) GetEncompassingSection(t int64) *section {
	// do a binary search to find the correct section
	left, right := 0, len(c.sections)-1
	for left <= right {
		mid := (left + right) / 2
		section := c.sections[mid]
		if section.from.Tick <= t && section.to.Tick >= t {
			// jackpot
			return &section
		}

		// adjust bounds based on comparison
		if section.to.Tick < t {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return nil
}
