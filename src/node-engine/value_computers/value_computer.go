package valuecomputers

import (
	"fmt"

	opcnode "github.com/AndreiLacatos/opc-engine/node-engine/models/opc/opc_node"
	"github.com/AndreiLacatos/opc-engine/node-engine/models/waveform"
	waveformvalue "github.com/AndreiLacatos/opc-engine/node-engine/models/waveform/waveform_value"
	"go.uber.org/zap"
)

type ValueComputer interface {
	Init()
	GetValueAtTick(t int64) waveformvalue.WaveformPointValue
}

func MakeValueComputer(n opcnode.OpcValueNode, l *zap.Logger) *ValueComputer {
	log := l.Named("VALCOMP")
	switch n.Waveform.WaveformType {
	case waveform.Transitions:
		return makeTransitionValueComputer(n, log)
	case waveform.NumericValues:
		return makeNumericValueComputer(n, log)
	}

	log.Warn(fmt.Sprintf("unrecognized waveform type %v", n.Waveform.WaveformType))
	return nil
}

func makeTransitionValueComputer(n opcnode.OpcValueNode, l *zap.Logger) *ValueComputer {
	var c ValueComputer = &transitionStrategyCalculator{
		logger:   l,
		waveform: n.Waveform,
	}
	return &c
}

func makeNumericValueComputer(n opcnode.OpcValueNode, l *zap.Logger) *ValueComputer {
	if meta, ok := (*n.Waveform.Meta).(waveform.NumericWaveformMeta); !ok {
		l.Warn(fmt.Sprintf("invalid waveform meta for %s", opcnode.ToDebugString(&n)))
		return nil
	} else {
		switch meta.Smoothing {
		case waveform.Step:
			var c ValueComputer = &stepSmoothingStrategyCalculator{
				logger:   l,
				waveform: n.Waveform,
			}
			return &c
		case waveform.Linear:
			var c ValueComputer = &linearSmoothingStrategyCalculator{
				logger:   l,
				waveform: n.Waveform,
			}
			return &c
		case waveform.CubicSpline:
			var c ValueComputer = &cubicSplineSmoothingStrategyCalculator{
				logger:   l,
				waveform: n.Waveform,
			}
			return &c
		default:
			l.Warn(fmt.Sprintf("unrecognized smoothing strategy %v for %s", meta.Smoothing, opcnode.ToDebugString(&n)))
			return nil
		}
	}
}
