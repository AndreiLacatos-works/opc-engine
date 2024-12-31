package serialization

import (
	"log"

	"github.com/AndreiLacatos/opc-engine/node-engine/models/opc"
	opcnode "github.com/AndreiLacatos/opc-engine/node-engine/models/opc/opc_node"
	"github.com/google/uuid"
)

type OpcStructureModel struct {
	Root OpcStructureNodeModel `json:"root"`
}

type OpcStructureNodeModel struct {
	Id       string                   `json:"id"`
	Label    string                   `json:"label"`
	NodeType string                   `json:"type"`
	Children *[]OpcStructureNodeModel `json:"children,omitempty"`
	Waveform *WaveformModel           `json:"waveform,omitempty"`
}

func (m *OpcStructureModel) ToDomain() opc.OpcStructure {
	return opc.OpcStructure{
		Root: *m.Root.ToDomain().(*opcnode.OpcContainerNode),
	}
}

func (n *OpcStructureNodeModel) ToDomain() opcnode.OpcStructureNode {
	id, err := uuid.Parse(n.Id)
	if err != nil {
		log.Printf("%s is not a valid UUID, skipping node\n", n.Id)
		return nil
	}
	switch n.NodeType {
	case "container":
		mappedChildren := make([]opcnode.OpcStructureNode, 0)
		for _, v := range *n.Children {
			if mapped := v.ToDomain(); mapped != nil {
				mappedChildren = append(mappedChildren, mapped)
			}
		}
		return &opcnode.OpcContainerNode{
			Id:       id,
			Label:    n.Label,
			Children: mappedChildren,
		}
	case "value":
		return &opcnode.OpcValueNode{
			Id:       id,
			Label:    n.Label,
			Waveform: n.Waveform.ToDomain(),
		}
	default:
		log.Printf("unrecognized node type %s, skipping node\n", n.NodeType)
		return nil
	}
}
