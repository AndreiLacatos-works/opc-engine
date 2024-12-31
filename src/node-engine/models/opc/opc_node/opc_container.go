package opcnode

type OpcContainerNode struct {
	Id string
	Label string
	Children [] OpcStructureNode
}

func (c *OpcContainerNode) GetId() string {
	return c.Id
}

func (c *OpcContainerNode) GetLabel() string {
	return c.Label
}
