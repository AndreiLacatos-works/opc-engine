package serialization

import opcserialization "github.com/AndreiLacatos/opc-engine/node-engine/serialization"

type Command struct {
	Command string                             `json:"command"`
	Payload opcserialization.OpcStructureModel `json:"payload"`
}

type Respose struct {
	Status string  `json:"status"`
	Reason *string `json:"reason"`
}
