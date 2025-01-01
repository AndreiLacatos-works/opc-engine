package opcnode

import (
	"fmt"

	"github.com/google/uuid"
)

type OpcStructureNode interface {
	GetId() uuid.UUID
	GetLabel() string
}

func ToDebugString(v OpcStructureNode) string {
	return fmt.Sprintf("%s (%s)", v.GetLabel(), v.GetId())
}
