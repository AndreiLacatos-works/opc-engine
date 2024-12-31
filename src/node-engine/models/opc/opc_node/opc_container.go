package opcnode

import "github.com/google/uuid"

type OpcContainerNode struct {
	Id       uuid.UUID
	Label    string
	Children []OpcStructureNode
}

func (c *OpcContainerNode) GetId() uuid.UUID {
	return c.Id
}

func (c *OpcContainerNode) GetLabel() string {
	return c.Label
}
