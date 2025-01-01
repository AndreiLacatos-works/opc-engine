package nodeengine

import (
	"context"
	"fmt"
	"time"

	opcnode "github.com/AndreiLacatos/opc-engine/node-engine/models/opc/opc_node"
	"go.uber.org/zap"
)

type valueChangeEngineImpl struct {
	Nodes  []opcnode.OpcValueNode
	Cancel context.CancelFunc
	Events chan NodeValueChange
	Logger *zap.Logger
}

func (e *valueChangeEngineImpl) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	e.Cancel = cancel
	for _, n := range e.Nodes {
		go e.executeEngineLoop(ctx, n)
	}
}

func (e *valueChangeEngineImpl) executeEngineLoop(ctx context.Context, n opcnode.OpcValueNode) {
	e.Logger.Info(fmt.Sprintf("starting engine loop for %s", n.Label))
	for {
		previousTick := int64(0)
		transitions := n.Waveform.TransitionPoints
		for _, transition := range transitions {
			delta := transition.Tick - previousTick
			select {
			case <-ctx.Done():
				e.Logger.Debug(fmt.Sprintf("engine loop done for %s", n.Label))
				return
			case <-time.After(time.Duration(delta) * time.Millisecond):
				defer func() {
					if r := recover(); r != nil {
						e.Logger.Debug("attempted to push value change but event channel was closed")
					}
				}()
				e.Logger.Debug(fmt.Sprintf("emitting new value %f for %s",
					transition.Value.GetValue(), opcnode.ToDebugString(&n)))
				e.Events <- NodeValueChange{
					Node:     n,
					NewValue: transition.Value,
				}
			}
			previousTick = transition.Tick
		}

		lastTransitionTick := transitions[len(transitions)-1].Tick
		untilEnd := n.Waveform.Duration - lastTransitionTick
		select {
		case <-ctx.Done():
			e.Logger.Info(fmt.Sprintf("engine loop done for %s", n.Label))
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
}
