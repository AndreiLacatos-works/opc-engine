package nodeengine_test

import (
	"context"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	nodeengine "github.com/AndreiLacatos/opc-engine/node-engine"
	"github.com/AndreiLacatos/opc-engine/node-engine/models/opc"
	opcnode "github.com/AndreiLacatos/opc-engine/node-engine/models/opc/opc_node"
	"github.com/AndreiLacatos/opc-engine/node-engine/models/waveform"
	waveformvalue "github.com/AndreiLacatos/opc-engine/node-engine/models/waveform/waveform_value"
	"github.com/google/uuid"
	"go.uber.org/zap"
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
	nodeSamples := c.CollectSamples(context.TODO(), e, time.Duration(int(float32(n.Waveform.Duration)*0.95))*time.Millisecond)

	// assert
	booleanSamples := nodeSamples[n.Id].samples
	expectedSamples, err := loadExpectedBooleanResultSet("boolean test data.csv")
	if err != nil {
		t.Errorf("could not load expected test results: %v", err)
		t.FailNow()
	}

	wiggle := time.Duration(3) * time.Millisecond
	adjustExpectedTimestamps(&expectedSamples, testStart)
	printSamples(l, expectedSamples.samples, testStart)
	printSamples(l, booleanSamples, testStart)
	assertBooleanSamplesets(t, expectedSamples.samples, booleanSamples, wiggle)
}

func TestSingleBooleanNodeValues_TransitionEvery137Ms_WaveformDuration1700Ms_CollectionDuration4420Ms(t *testing.T) {
	// arrange
	l := zaptest.NewLogger(t)
	n := &opcnode.OpcValueNode{
		Id:    uuid.MustParse("da858518-50c9-4e55-b312-6370275b412d"),
		Label: "Boolean",
		Waveform: waveform.Waveform{
			Duration:      1700,
			TickFrequency: 137,
			WaveformType:  waveform.Transitions,
			TransitionPoints: []waveform.WaveformValue{
				{
					Tick:  137,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  274,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  411,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  548,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  685,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  822,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  959,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  1096,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  1233,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  1370,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  1507,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  1644,
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
	nodeSamples := c.CollectSamples(context.TODO(), e, time.Duration(4420)*time.Millisecond)

	// assert
	booleanSamples := nodeSamples[n.Id].samples
	expectedSamples, err := loadExpectedBooleanResultSet("boolean test data 2.csv")
	if err != nil {
		t.Errorf("could not load expected test results: %v", err)
		t.FailNow()
	}

	wiggle := time.Duration(3) * time.Millisecond
	adjustExpectedTimestamps(&expectedSamples, testStart)
	printSamples(l, expectedSamples.samples, testStart)
	printSamples(l, booleanSamples, testStart)
	assertBooleanSamplesets(t, expectedSamples.samples, booleanSamples, wiggle)
}

func TestSingleBooleanNodeValues_TransitionsRandomly_WaveformDuration1300Ms_CollectionDuration6140Ms(t *testing.T) {
	// arrange
	l := zaptest.NewLogger(t)
	n := &opcnode.OpcValueNode{
		Id:    uuid.MustParse("da858518-50c9-4e55-b312-6370275b412d"),
		Label: "Boolean",
		Waveform: waveform.Waveform{
			Duration:      1300,
			TickFrequency: 75,
			WaveformType:  waveform.Transitions,
			TransitionPoints: []waveform.WaveformValue{
				{
					Tick:  150,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  225,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  450,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  825,
					Value: &waveformvalue.Transition{},
				},
				{
					Tick:  900,
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
	nodeSamples := c.CollectSamples(context.TODO(), e, time.Duration(6140)*time.Millisecond)

	// assert
	booleanSamples := nodeSamples[n.Id].samples
	expectedSamples, err := loadExpectedBooleanResultSet("boolean test data 3.csv")
	if err != nil {
		t.Errorf("could not load expected test results: %v", err)
		t.FailNow()
	}

	wiggle := time.Duration(3) * time.Millisecond
	adjustExpectedTimestamps(&expectedSamples, testStart)
	printSamples(l, expectedSamples.samples, testStart)
	printSamples(l, booleanSamples, testStart)
	assertBooleanSamplesets(t, expectedSamples.samples, booleanSamples, wiggle)
}

func TestSingleNumericNodeValues_StepSmoothing_WaveformDuration2700Ms_CollectionDuration6185Ms(t *testing.T) {
	// arrange
	l := zaptest.NewLogger(t)
	var m waveform.WaveformMeta = waveform.NumericWaveformMeta{
		Smoothing: waveform.Step,
	}
	n := &opcnode.OpcValueNode{
		Id:    uuid.MustParse("da858518-50c9-4e55-b312-6370275b412d"),
		Label: "Numbers",
		Waveform: waveform.Waveform{
			Duration:      2700,
			TickFrequency: 50,
			WaveformType:  waveform.NumericValues,
			Meta:          &m,
			TransitionPoints: []waveform.WaveformValue{
				{
					Tick: 250,
					Value: &waveformvalue.DoubleValue{
						Value: 40.0,
					},
				},
				{
					Tick: 500,
					Value: &waveformvalue.DoubleValue{
						Value: 18.42,
					},
				},
				{
					Tick: 1000,
					Value: &waveformvalue.DoubleValue{
						Value: -14.75,
					},
				},
				{
					Tick: 1500,
					Value: &waveformvalue.DoubleValue{
						Value: 0.17,
					},
				},
				{
					Tick: 1650,
					Value: &waveformvalue.DoubleValue{
						Value: 10.57,
					},
				},
				{
					Tick: 2250,
					Value: &waveformvalue.DoubleValue{
						Value: 4.8,
					},
				},
				{
					Tick: 2400,
					Value: &waveformvalue.DoubleValue{
						Value: 69.02,
					},
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
	nodeSamples := c.CollectSamples(context.TODO(), e, time.Duration(6185)*time.Millisecond)

	// assert
	numericSamples := nodeSamples[n.Id].samples
	expectedSamples, err := loadExpectedNumericResultSet("numeric values - steps.csv")
	if err != nil {
		t.Errorf("could not load expected test results: %v", err)
		t.FailNow()
	}

	wiggle := time.Duration(3) * time.Millisecond
	adjustExpectedTimestamps(&expectedSamples, testStart)
	printSamples(l, expectedSamples.samples, testStart)
	printSamples(l, numericSamples, testStart)
	assertNumericSamplesets(t, expectedSamples.samples, numericSamples, wiggle)
}

func TestSingleNumericNodeValues_LinearSmoothing_WaveformDuration2700Ms_CollectionDuration6185Ms(t *testing.T) {
	// arrange
	l := zaptest.NewLogger(t)
	var m waveform.WaveformMeta = waveform.NumericWaveformMeta{
		Smoothing: waveform.Linear,
	}
	n := &opcnode.OpcValueNode{
		Id:    uuid.MustParse("da858518-50c9-4e55-b312-6370275b412d"),
		Label: "Numbers",
		Waveform: waveform.Waveform{
			Duration:      2700,
			TickFrequency: 50,
			WaveformType:  waveform.NumericValues,
			Meta:          &m,
			TransitionPoints: []waveform.WaveformValue{
				{
					Tick: 250,
					Value: &waveformvalue.DoubleValue{
						Value: 40.0,
					},
				},
				{
					Tick: 500,
					Value: &waveformvalue.DoubleValue{
						Value: 18.42,
					},
				},
				{
					Tick: 1000,
					Value: &waveformvalue.DoubleValue{
						Value: -14.75,
					},
				},
				{
					Tick: 1500,
					Value: &waveformvalue.DoubleValue{
						Value: 0.17,
					},
				},
				{
					Tick: 1650,
					Value: &waveformvalue.DoubleValue{
						Value: 10.57,
					},
				},
				{
					Tick: 2250,
					Value: &waveformvalue.DoubleValue{
						Value: 4.8,
					},
				},
				{
					Tick: 2400,
					Value: &waveformvalue.DoubleValue{
						Value: 69.02,
					},
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
	nodeSamples := c.CollectSamples(context.TODO(), e, time.Duration(6185)*time.Millisecond)

	// assert
	numericSamples := nodeSamples[n.Id].samples
	expectedSamples, err := loadExpectedNumericResultSet("numeric values - linear.csv")
	if err != nil {
		t.Errorf("could not load expected test results: %v", err)
		t.FailNow()
	}

	wiggle := time.Duration(3) * time.Millisecond
	adjustExpectedTimestamps(&expectedSamples, testStart)
	printSamples(l, expectedSamples.samples, testStart)
	printSamples(l, numericSamples, testStart)
	assertNumericSamplesets(t, expectedSamples.samples, numericSamples, wiggle)
}

func TestSingleNumericNodeValues_CubicSplineSmoothing_WaveformDuration2700Ms_CollectionDuration6185Ms(t *testing.T) {
	// arrange
	l := zaptest.NewLogger(t)
	var m waveform.WaveformMeta = waveform.NumericWaveformMeta{
		Smoothing: waveform.CubicSpline,
	}
	n := &opcnode.OpcValueNode{
		Id:    uuid.MustParse("da858518-50c9-4e55-b312-6370275b412d"),
		Label: "Numbers",
		Waveform: waveform.Waveform{
			Duration:      2700,
			TickFrequency: 50,
			WaveformType:  waveform.NumericValues,
			Meta:          &m,
			TransitionPoints: []waveform.WaveformValue{
				{
					Tick: 250,
					Value: &waveformvalue.DoubleValue{
						Value: 40.0,
					},
				},
				{
					Tick: 500,
					Value: &waveformvalue.DoubleValue{
						Value: 18.42,
					},
				},
				{
					Tick: 1000,
					Value: &waveformvalue.DoubleValue{
						Value: -14.75,
					},
				},
				{
					Tick: 1500,
					Value: &waveformvalue.DoubleValue{
						Value: 0.17,
					},
				},
				{
					Tick: 1650,
					Value: &waveformvalue.DoubleValue{
						Value: 10.57,
					},
				},
				{
					Tick: 2250,
					Value: &waveformvalue.DoubleValue{
						Value: 4.8,
					},
				},
				{
					Tick: 2400,
					Value: &waveformvalue.DoubleValue{
						Value: 69.02,
					},
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
	nodeSamples := c.CollectSamples(context.TODO(), e, time.Duration(6185)*time.Millisecond)

	// assert
	numericSamples := nodeSamples[n.Id].samples
	expectedSamples, err := loadExpectedNumericResultSet("numeric values - cubic spline.csv")
	if err != nil {
		t.Errorf("could not load expected test results: %v", err)
		t.FailNow()
	}

	wiggle := time.Duration(3) * time.Millisecond
	adjustExpectedTimestamps(&expectedSamples, testStart)
	printSamples(l, expectedSamples.samples, testStart)
	printSamples(l, numericSamples, testStart)
	assertNumericSamplesets(t, expectedSamples.samples, numericSamples, wiggle)
}

func areClose(t1, t2 time.Time, wiggleRoom time.Duration) bool {
	diff := t1.Sub(t2)
	return diff <= wiggleRoom && diff >= -wiggleRoom
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.999")
}

func loadExpectedBooleanResultSet(f string) (ResultSet, error) {
	r := ResultSet{
		samples: make([]Sample, 0),
	}

	file, err := os.Open(path.Join("..", "..", "test data", f))
	if err != nil {
		return r, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	rows, err := reader.ReadAll()
	if err != nil {
		return r, err
	}

	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) != 2 {
			continue
		}

		number, err := strconv.Atoi(row[0])
		if err != nil {
			continue
		}

		val, err := strconv.ParseBool(row[1])
		if err != nil {
			continue
		}

		var t time.Time
		r.samples = append(r.samples, Sample{
			timestamp: t.Add(time.Duration(number) * time.Millisecond),
			value: &waveformvalue.Transition{
				Value: val,
			},
		})
	}

	return r, nil
}

func loadExpectedNumericResultSet(f string) (ResultSet, error) {
	r := ResultSet{
		samples: make([]Sample, 0),
	}

	file, err := os.Open(path.Join("..", "..", "test data", f))
	if err != nil {
		return r, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	rows, err := reader.ReadAll()
	if err != nil {
		return r, err
	}

	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) != 2 {
			continue
		}

		number, err := strconv.Atoi(row[0])
		if err != nil {
			continue
		}

		val, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			continue
		}

		var t time.Time
		r.samples = append(r.samples, Sample{
			timestamp: t.Add(time.Duration(number) * time.Millisecond),
			value: &waveformvalue.DoubleValue{
				Value: val,
			},
		})
	}

	return r, nil
}

func adjustExpectedTimestamps(r *ResultSet, o time.Time) {
	var t time.Time
	for i := range r.samples {
		delta := r.samples[i].timestamp.Sub(t)
		r.samples[i].timestamp = o.Add(delta)
	}
}

func assertBooleanSamplesets(t *testing.T, expected []Sample, actual []Sample, wiggle time.Duration) {
	if len(actual) != len(expected) {
		t.Errorf("expected %d samples and got %d", len(expected), len(actual))
		t.FailNow()
	}

	for i, e := range expected {
		if !areClose(actual[i].timestamp, e.timestamp, wiggle) {
			t.Errorf("expected sample %d to take place on %s, actual: %s", i+1, formatDate(e.timestamp), formatDate(actual[i].timestamp))
			t.FailNow()
		}
		if actual[i].value.GetValue() != e.value.GetValue() {
			t.Errorf("at %s expected value %v, actual: %s", formatDate(e.timestamp), e.value.GetValue(), actual[i].value.GetValue())
			t.FailNow()
		}
	}
}

func assertNumericSamplesets(t *testing.T, expected []Sample, actual []Sample, wiggle time.Duration) {
	if len(actual) != len(expected) {
		t.Errorf("expected %d samples and got %d", len(expected), len(actual))
		t.FailNow()
	}

	for i, e := range expected {
		if !areClose(actual[i].timestamp, e.timestamp, wiggle) {
			t.Errorf("expected sample %d to take place on %s, actual: %s", i+1, formatDate(e.timestamp), formatDate(actual[i].timestamp))
			t.FailNow()
		}
		expectedValue := e.value.GetValue().(float64)
		actualValue := actual[i].value.GetValue().(float64)
		diff := expectedValue - actualValue
		if math.Abs(diff) > 0.02 {
			t.Errorf("at %s expected value %v, actual: %s", formatDate(e.timestamp), e.value.GetValue(), actual[i].value.GetValue())
			t.FailNow()
		}
	}
}

func printSamples(l *zap.Logger, s []Sample, o time.Time) {
	l.Info("Samples (time,tick,value):")
	for _, e := range s {
		l.Info(fmt.Sprintf("%s,%d,%v", formatDate(e.timestamp), e.timestamp.Sub(o)/1e6, e.value.GetValue()))
	}
}
