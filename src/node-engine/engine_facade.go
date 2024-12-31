package nodeengine

import (
	opcnode "github.com/AndreiLacatos/opc-engine/node-engine/models/opc/opc_node"
	waveformvalue "github.com/AndreiLacatos/opc-engine/node-engine/models/waveform/waveform_value"
	"github.com/google/uuid"
)

type NodeValueChange struct {
	Id       uuid.UUID
	NewValue waveformvalue.WaveformPointValue
}

type ValueChangeEngine interface {
	Start()
	EventChannel() chan NodeValueChange
	Stop()
}

func CreateNew(nodes []opcnode.OpcValueNode) ValueChangeEngine {
	return &valueChangeEngineImpl{Nodes: nodes, Events: make(chan NodeValueChange)}
}
