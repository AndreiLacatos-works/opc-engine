package valuecomputers

import (
	"fmt"

	"github.com/AndreiLacatos/opc-engine/node-engine/models/waveform"
	waveformvalue "github.com/AndreiLacatos/opc-engine/node-engine/models/waveform/waveform_value"
	"go.uber.org/zap"
)

type coefficients struct {
	a, b, c, d []float64
}

type cubicSplineSmoothingStrategyCalculator struct {
	logger       *zap.Logger
	waveform     waveform.Waveform
	coefficients coefficients
	x            []float64
}

func (c *cubicSplineSmoothingStrategyCalculator) Init() {
	l := len(c.waveform.TransitionPoints)
	if l < 3 {
		c.logger.Warn(fmt.Sprintf("can not use cubic spline smoothing strategy for %d transition points", l))
		c.coefficients = coefficients{
			a: make([]float64, 5),
			b: make([]float64, 5),
			c: make([]float64, 5),
			d: make([]float64, 5),
		}
		c.x = make([]float64, 5)
		return
	}

	// map the explicit transition points to two arrays holding
	// ticks for X & value for Y coordinates
	x := make([]float64, len(c.waveform.TransitionPoints))
	y := make([]float64, len(c.waveform.TransitionPoints))
	for i, v := range c.waveform.TransitionPoints {
		x[i] = float64(v.Tick)
		y[i] = v.Value.GetValue().(float64)
	}

	// add two additional entries for X = 0 & X = waveform.Duration
	if x[0] != 0 {
		x = append([]float64{0}, x...)
		y = append([]float64{c.waveform.TransitionPoints[0].Value.GetValue().(float64)}, y...)
	}
	if x[len(x)-1] != float64(c.waveform.Duration) {
		x = append(x, float64(c.waveform.Duration))
		y = append(y, c.waveform.TransitionPoints[l-1].Value.GetValue().(float64))
	}

	ca, cb, cc, cd := computeCubicSplineCoefficients(x, y)
	c.coefficients = coefficients{
		a: ca,
		b: cb,
		c: cc,
		d: cd,
	}
	c.x = x
}

func (c *cubicSplineSmoothingStrategyCalculator) GetValueAtTick(t int64) waveformvalue.WaveformPointValue {
	return &waveformvalue.DoubleValue{
		Value: c.interpolate(t),
	}
}

func computeCubicSplineCoefficients(x []float64, y []float64) (a, b, c, d []float64) {
	n := len(x) - 1

	// step 1: calculate the h's, the differences between adjacent x's
	h := make([]float64, n)
	for i := 0; i < n; i++ {
		h[i] = x[i+1] - x[i]
	}

	// step 2: set up the alpha array
	alpha := make([]float64, n)
	for i := 1; i < n; i++ {
		alpha[i] = (3/h[i])*(y[i+1]-y[i]) - (3/h[i-1])*(y[i]-y[i-1])
	}

	// step 3: set up the matrix system for the c coefficients (tridiagonal system)
	l := make([]float64, n+1)
	mu := make([]float64, n)
	z := make([]float64, n+1)

	l[0] = 1.0
	for i := 1; i < n; i++ {
		l[i] = 2*(x[i+1]-x[i-1]) - h[i-1]*mu[i-1]
		mu[i] = h[i] / l[i]
		z[i] = (alpha[i] - h[i-1]*z[i-1]) / l[i]
	}

	l[n] = 1.0
	z[n] = 0.0

	// step 4: back substitution to calculate c coefficients
	c = make([]float64, n+1)
	for j := n - 1; j >= 0; j-- {
		c[j] = z[j] - mu[j]*c[j+1]
	}

	// step 5: calculate b and d coefficients
	b = make([]float64, n)
	d = make([]float64, n)
	a = y

	for i := 0; i < n; i++ {
		b[i] = (y[i+1]-y[i])/h[i] - h[i]*(c[i+1]+2*c[i])/3
		d[i] = (c[i+1] - c[i]) / (3 * h[i])
	}

	return a, b, c, d
}

func (c *cubicSplineSmoothingStrategyCalculator) interpolate(t int64) float64 {
	xQuery := float64(t)
	n := len(c.x) - 1

	// find the interval containing xQuery
	var i int
	for i = 0; i < n; i++ {
		if xQuery <= c.x[i+1] {
			break
		}
	}

	dx := xQuery - c.x[i]
	ca := c.coefficients.a
	cb := c.coefficients.b
	cc := c.coefficients.c
	cd := c.coefficients.d
	return ca[i] + cb[i]*dx + cc[i]*dx*dx + cd[i]*dx*dx*dx
}
