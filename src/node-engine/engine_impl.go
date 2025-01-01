package nodeengine

import (
	"context"
	"log"
	"time"

	opcnode "github.com/AndreiLacatos/opc-engine/node-engine/models/opc/opc_node"
)

type valueChangeEngineImpl struct {
	Nodes  []opcnode.OpcValueNode
	Cancel context.CancelFunc
	Events chan NodeValueChange
}

func (e *valueChangeEngineImpl) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	e.Cancel = cancel
	for _, n := range e.Nodes {
		go e.executeEngineLoop(ctx, n)
	}
}

func (e *valueChangeEngineImpl) executeEngineLoop(ctx context.Context, n opcnode.OpcValueNode) {
	log.Printf("starting engine loop for %s\n", n.Label)
	for {
		previousTick := int64(0)
		for _, transition := range n.Waveform.TransitionPoints {
			delta := transition.Tick - previousTick
			select {
			case <-ctx.Done():
				log.Printf("engine loop done for %s\n", n.Label)
				return
			case <-time.After(time.Duration(delta) * time.Millisecond):
				defer func() {
					if r := recover(); r != nil {
						log.Println("attempted to push value change but event channel was closed")
					}
				}()
				e.Events <- NodeValueChange{
					Node:     n,
					NewValue: transition.Value,
				}
			}
			previousTick = transition.Tick
		}

		untilEnd := n.Waveform.Duration - n.Waveform.TransitionPoints[len(n.Waveform.TransitionPoints)-1].Tick
		select {
		case <-ctx.Done():
			log.Printf("engine loop done for %s\n", n.Label)
			return
		case <-time.After(time.Duration(untilEnd) * time.Millisecond):
		}
	}
}

func (e *valueChangeEngineImpl) EventChannel() chan NodeValueChange {
	return e.Events
}

func (e *valueChangeEngineImpl) Stop() {
	log.Println("stopping value change engine")
	if e.Cancel != nil {
		e.Cancel()
	}
	close(e.Events)
}
