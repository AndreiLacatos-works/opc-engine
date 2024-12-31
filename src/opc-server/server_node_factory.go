package opcserver

import (
	"fmt"
	"time"

	opcnode "github.com/AndreiLacatos/opc-engine/node-engine/models/opc/opc_node"
	"github.com/AndreiLacatos/opc-engine/node-engine/models/waveform"
	"github.com/awcullen/opcua/server"
	"github.com/awcullen/opcua/ua"
)

func makeNode(r opcnode.OpcStructureNode, p ua.NodeID, s *server.Server) (server.Node, error) {
	switch t := r.(type) {
	case *opcnode.OpcContainerNode:
		return makeContainerNode(*t, p, s)
	case *opcnode.OpcValueNode:
		return makeValueNode(*t, p, s)
	default:
		return nil, fmt.Errorf("unsupported node type %v", t)
	}
}

func makeContainerNode(n opcnode.OpcContainerNode, p ua.NodeID, s *server.Server) (server.Node, error) {
	return server.NewObjectNode(
		s,
		ua.NewNodeIDGUID(2, n.GetId()),
		ua.NewQualifiedName(2, n.GetLabel()),
		ua.NewLocalizedText(n.GetLabel(), ""),
		ua.NewLocalizedText("", ""),
		nil,
		[]ua.Reference{
			// this entry links this object as a child to the parent
			{
				ReferenceTypeID: ua.NewNodeIDNumeric(0, 35),
				IsInverse:       true,
				TargetID:        ua.NewExpandedNodeID(p),
			},
			// this entry make this object a "folder" type
			{
				ReferenceTypeID: ua.NewNodeIDNumeric(0, 40),
				IsInverse:       false,
				TargetID:        ua.NewExpandedNodeID(ua.NewNodeIDNumeric(0, 61)),
			},
		},
		0,
	), nil
}

func makeValueNode(n opcnode.OpcValueNode, p ua.NodeID, s *server.Server) (server.Node, error) {
	nodeIdMap := map[waveform.WaveformType]ua.NodeID{
		waveform.Transitions:   ua.NewNodeIDNumeric(0, 1),
		waveform.NumericValues: ua.NewNodeIDNumeric(0, 11),
	}
	defaultValueMap := map[waveform.WaveformType]ua.Variant{
		waveform.Transitions:   false,
		waveform.NumericValues: float64(0),
	}
	typeNodeId, found := nodeIdMap[n.Waveform.WaveformType]
	if !found {
		return nil, fmt.Errorf("invalid waveform type %v", n.Waveform.WaveformType)
	}
	defaultValue, found := defaultValueMap[n.Waveform.WaveformType]
	if !found {
		return nil, fmt.Errorf("invalid waveform type %v", n.Waveform.WaveformType)
	}

	return server.NewVariableNode(
		s,
		ua.NewNodeIDGUID(2, n.GetId()),
		ua.NewQualifiedName(2, n.GetLabel()),
		ua.NewLocalizedText(n.GetLabel(), ""),
		ua.NewLocalizedText("", ""),
		nil,
		[]ua.Reference{
			{
				ReferenceTypeID: ua.NewNodeIDNumeric(0, 35),
				IsInverse:       true,
				TargetID:        ua.NewExpandedNodeID(p),
			},
		},
		ua.NewDataValue(defaultValue, ua.Good, time.Now().UTC(), 0, time.Now().UTC(), 0),
		typeNodeId,
		-1,
		[]uint32{},
		3,
		0,
		false,
		nil,
	), nil
}
