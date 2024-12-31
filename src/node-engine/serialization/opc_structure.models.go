package serialization

import (
	"fmt"

	opcnode "github.com/AndreiLacatos/opc-engine/node-engine/models/opc/opc_node"
)

type OpcStructureModel struct {
	Root OpcStructureNodeModel `json:"root"`
}


type OpcStructureNodeModel struct {
    Id        string				   `json:"id"`
    Label     string                   `json:"label"`
    NodeType  string                   `json:"type"`
    Children  *[]OpcStructureNodeModel `json:"children,omitempty"`
    Waveform  *WaveformModel           `json:"waveform,omitempty"`
}

func (n *OpcStructureNodeModel) ToDomain() opcnode.OpcStructureNode {
    switch (n.NodeType) {
    case "container": 
    mappedChildren := make([]opcnode.OpcStructureNode, len(*n.Children))
    for i, v := range *n.Children {
        mappedChildren[i] = v.ToDomain()
    }
    return &opcnode.OpcContainerNode{
        Id: n.Id,
        Label: n.Label,
        Children: mappedChildren,
    }
    case "value": return &opcnode.OpcValueNode{
        Id: n.Id,
        Label: n.Label,
        Waveform: n.Waveform.ToDomain(),
    }
    default: panic(fmt.Sprintf("unrecognized node type %s", n.NodeType))
    }
}
