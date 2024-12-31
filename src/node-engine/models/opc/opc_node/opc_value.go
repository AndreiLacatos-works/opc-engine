package opcnode

import "github.com/AndreiLacatos/opc-engine/node-engine/models/waveform"

type OpcValueNode struct {
	Id string
	Label string
	Waveform waveform.Waveform
}

func (v *OpcValueNode) GetId() string {
	return v.Id
}

func (v *OpcValueNode) GetLabel() string {
	return v.Label
}