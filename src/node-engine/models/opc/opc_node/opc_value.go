package opcnode

import (
	"github.com/AndreiLacatos/opc-engine/node-engine/models/waveform"
	"github.com/google/uuid"
)

type OpcValueNode struct {
	Id       uuid.UUID
	Label    string
	Waveform waveform.Waveform
}

func (v *OpcValueNode) GetId() uuid.UUID {
	return v.Id
}

func (v *OpcValueNode) GetLabel() string {
	return v.Label
}
