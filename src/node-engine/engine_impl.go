package nodeengine

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	opcnode "github.com/AndreiLacatos/opc-engine/node-engine/models/opc/opc_node"
	waveformvalue "github.com/AndreiLacatos/opc-engine/node-engine/models/waveform/waveform_value"
	valuecomputers "github.com/AndreiLacatos/opc-engine/node-engine/value_computers"
	"go.uber.org/zap"
)

type valueChangeEngineImpl struct {
	Nodes        []opcnode.OpcValueNode
	Cancel       context.CancelFunc
	Events       chan NodeValueChange
	Logger       *zap.Logger
	DebugEnabled bool
	Teardown     *sync.WaitGroup
}

func (e *valueChangeEngineImpl) Start() {
	e.Logger.Info("starting node engine")
	ctx, cancel := context.WithCancel(context.Background())
	e.Teardown = &sync.WaitGroup{}
	e.Cancel = cancel
	for _, n := range e.Nodes {
		go e.executeEngineLoop(ctx, n)
	}
}

func (e *valueChangeEngineImpl) executeEngineLoop(ctx context.Context, n opcnode.OpcValueNode) {
	e.Logger.Info(fmt.Sprintf("starting engine loop for %s", n.Label))
	e.Teardown.Add(1)
	c := valuecomputers.MakeValueComputer(n, e.Logger)

	if c == nil {
		e.Logger.Error(fmt.Sprintf("failed to generate value computer for %s, quitting engine loop", opcnode.ToDebugString(&n)))
		return
	}

	(*c).Init()
	tickCount := n.Waveform.Duration / int64(n.Waveform.TickFrequency)

	for {
		for i := int64(0); i < tickCount; i++ {
			// emit value for current tick
			t := i * int64(n.Waveform.TickFrequency)
			v := (*c).GetValueAtTick(t)
			e.debugWrite(t, v)
			defer func() {
				if r := recover(); r != nil {
					e.Logger.Debug("attempted to push value change but event channel was closed")
				}
			}()
			e.Logger.Debug(fmt.Sprintf("emitting new value %f for %s", v.GetValue(), opcnode.ToDebugString(&n)))
			e.Events <- NodeValueChange{
				Node:     n,
				NewValue: v,
			}

			// wait for next tick
			select {
			case <-ctx.Done():
				e.Logger.Info(fmt.Sprintf("engine loop done for %s", n.Label))
				e.Teardown.Done()
				return
			case <-time.After(time.Duration(n.Waveform.TickFrequency) * time.Millisecond):
			}
		}

		// compute remaining time to complete waveform
		untilEnd := n.Waveform.Duration - int64(n.Waveform.TickFrequency)*tickCount
		select {
		case <-ctx.Done():
			e.Logger.Info(fmt.Sprintf("engine loop done for %s", n.Label))
			e.Teardown.Done()
			return
		case <-time.After(time.Duration(untilEnd) * time.Millisecond):
		}
	}
}

func (e *valueChangeEngineImpl) EventChannel() chan NodeValueChange {
	return e.Events
}

func (e *valueChangeEngineImpl) Stop() {
	e.Logger.Info("stopping value change engine")
	if e.Cancel != nil {
		e.Cancel()
	}
	close(e.Events)
	e.Teardown.Wait()
}

func (e *valueChangeEngineImpl) debugWrite(t int64, v waveformvalue.WaveformPointValue) {
	if !e.DebugEnabled {
		return
	}

	fileName := "debug.csv"
	lineToAppend := fmt.Sprintf("%d,%f\n", t, v.GetValue())

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		e.Logger.Debug(fmt.Sprintf("error opening debug file: %v\n", err))
		return
	}
	defer file.Close()

	_, err = file.WriteString(lineToAppend)
	if err != nil {
		e.Logger.Debug(fmt.Sprintf("error writing to debug file: %v\n", err))
		return
	}
}
