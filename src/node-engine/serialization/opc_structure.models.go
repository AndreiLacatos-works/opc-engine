package serialization

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
