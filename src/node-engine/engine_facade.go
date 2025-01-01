package nodeengine

import (
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

func CreateNew(nodes []opcnode.OpcValueNode, l *zap.Logger) ValueChangeEngine {
	return &valueChangeEngineImpl{
		Nodes:  nodes,
		Events: make(chan NodeValueChange),
		Logger: l.Named("ENGINE"),
	}
}
