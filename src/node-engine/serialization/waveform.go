package serialization

type WaveformModel struct {
	Duration int64 							`json:"duration"`
	TickFrequency int32 					`json:"tickFrequency"`
	WaveformType string 					`json:"type"`
	TransitionPoints []WaveformValueModel 	`json:"transitionPoints"`
}

type WaveformValueModel struct {
	Tick int64  	`json:"tick"`
	Value float64 	`json:"value"`
}
