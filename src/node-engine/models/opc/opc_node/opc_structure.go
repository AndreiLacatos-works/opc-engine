package opcnode

import "github.com/google/uuid"

type OpcStructureNode interface {
	GetId() uuid.UUID
	GetLabel() string
}
