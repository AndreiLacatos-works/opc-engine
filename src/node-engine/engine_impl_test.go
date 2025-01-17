package nodeengine_test

import (
	"context"
	"testing"
	"time"

	nodeengine "github.com/AndreiLacatos/opc-engine/node-engine"
	"github.com/AndreiLacatos/opc-engine/node-engine/models/opc"
	opcnode "github.com/AndreiLacatos/opc-engine/node-engine/models/opc/opc_node"
	"github.com/AndreiLacatos/opc-engine/node-engine/models/waveform"
	waveformvalue "github.com/AndreiLacatos/opc-engine/node-engine/models/waveform/waveform_value"
	"github.com/google/uuid"
	"go.uber.org/zap/zaptest"
)

type Sample struct {
	timestamp time.Time
	value     waveformvalue.WaveformPointValue
}

type ResultSet struct {
	samples []Sample
}

type SampleCollector struct {
	Ctx context.Context
	Acc map[uuid.UUID]ResultSet
}

func (c *SampleCollector) CollectSamples(ctx context.Context, e nodeengine.ValueChangeEngine, t time.Duration) map[uuid.UUID]ResultSet {
	c.Acc = make(map[uuid.UUID]ResultSet)
	var cancel context.CancelFunc
	c.Ctx, cancel = context.WithDeadline(ctx, time.Now().Add(t))

	go c.Subscribe(e)
	go e.Start()
	defer e.Stop()

	<-c.Ctx.Done()
	switch c.Ctx.Err() {
	case context.DeadlineExceeded:
		// normal case, sampling interval elapsed
		// stop collection & return results
		cancel()
		return c.Acc
	default:
		// context was canceled or some other error occurred
		cancel()
		return make(map[uuid.UUID]ResultSet)
	}
}

func (c *SampleCollector) Subscribe(e nodeengine.ValueChangeEngine) {
	ch := e.EventChannel()
	for {
		select {
		case <-c.Ctx.Done():
			// subsciption canceled
			return
		case v := <-ch:
			s := Sample{
				timestamp: time.Now(),
				value:     v.NewValue,
			}

			r, ok := c.Acc[v.Node.Id]
			if !ok {
				r = ResultSet{}
			}

			r.samples = append(r.samples, s)
			c.Acc[v.Node.Id] = r
		}
	}
}

func TestSingleBooleanNodeValues_TransitionEvery200Ms_SingleCycle(t *testing.T) {
	// arrange
	l := zaptest.NewLogger(t)
	n := &opcnode.OpcValueNode{
		Id:    uuid.MustParse("da858518-50c9-4e55-b312-6370275b412d"),
		Label: "Boolean",
		Waveform: waveform.Waveform{
			Duration:      2000,
			TickFrequency: 200,
			WaveformType:  waveform.Transitions,
			TransitionPoints: []waveform.WaveformValue{
				{
					Tick:  200,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  400,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  600,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  800,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  1000,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  1200,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  1400,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  1600,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  1800,
					Value: &waveformvalue.Transition{},
				},
			},
		},
	}
	s := opc.OpcStructure{
		Root: opcnode.OpcContainerNode{
			Id:    uuid.New(),
			Label: "Root",
			Children: []opcnode.OpcStructureNode{
				n,
			},
		},
	}
	e := nodeengine.CreateNew(s, l, false)
	c := SampleCollector{}

	// act
	testStart := time.Now()
	nodeSamples := c.CollectSamples(context.TODO(), e, time.Duration(n.Waveform.Duration)*time.Millisecond)

	// assert
	booleanSamples := nodeSamples[n.Id]

	// expect one sample for each transition point & an extra one for the very
	// first tick (at tick 0 as soon as the engine boots it emits the starting value)
	expectedSampleCount := len(n.Waveform.TransitionPoints) + 1
	expectedTransitionTimestamps := make([]time.Time, 0, expectedSampleCount)
	expectedTransitionTimestamps = append(expectedTransitionTimestamps, testStart)
	for _, t := range n.Waveform.TransitionPoints {
		delta := time.Duration(t.Tick) * time.Millisecond
		expectedTransitionTimestamps = append(expectedTransitionTimestamps, testStart.Add(delta))
	}

	if len(booleanSamples.samples) != expectedSampleCount {
		t.Errorf("expected %d samples and got %d", expectedSampleCount, len(booleanSamples.samples))
		t.FailNow()
	}

	sampleSet := nodeSamples[n.Id].samples

	wiggle := time.Duration(15) * time.Millisecond
	for i, e := range expectedTransitionTimestamps {
		if !areClose(sampleSet[i].timestamp, e, wiggle) {
			t.Errorf("expected sample %d to take place on %s, actual: %s", i+1, formatDate(e), formatDate(sampleSet[i].timestamp))
			t.FailNow()
		}
	}
}

func areClose(t1, t2 time.Time, wiggleRoom time.Duration) bool {
	diff := t1.Sub(t2)
	return diff <= wiggleRoom && diff >= -wiggleRoom
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.999")
}
