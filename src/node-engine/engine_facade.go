package nodeengine

import (
	"github.com/AndreiLacatos/opc-engine/node-engine/models/opc"
	opcnode "github.com/AndreiLacatos/opc-engine/node-engine/models/opc/opc_node"
	waveformvalue "github.com/AndreiLacatos/opc-engine/node-engine/models/waveform/waveform_value"
	"go.uber.org/zap"
)

type NodeValueChange struct {
	Node     opcnode.OpcValueNode
	NewValue waveformvalue.WaveformPointValue
}

type ValueChangeEngine interface {
	Start()
	EventChannel() chan NodeValueChange
	Stop()
}

func CreateNew(s opc.OpcStructure, l *zap.Logger, debug bool) ValueChangeEngine {
	return &valueChangeEngineImpl{
		Nodes:        extractValueNodes(s.Root),
		Events:       make(chan NodeValueChange),
		Logger:       l.Named("ENGINE"),
		DebugEnabled: debug,
	}
}

func extractValueNodes(r opcnode.OpcContainerNode) []opcnode.OpcValueNode {
	res := make([]opcnode.OpcValueNode, 0)

	for _, n := range r.Children {
		switch t := n.(type) {
		case *opcnode.OpcContainerNode:
			res = append(res, extractValueNodes(*t)...)
		case *opcnode.OpcValueNode:
			res = append(res, *t)
		}
	}
	return res
}
