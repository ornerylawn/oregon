package wavetable

import (
	"errors"
	"math"
)

const (
	twoPi = 2 * math.Pi
)

type table struct {
	samples         []float64
	sizeMask        int
	phaseTableRatio float64
}

type WaveTable struct {
	tables []*table
}

func New(amp []float64) (*WaveTable, error) {
	if len(amp) != 20000 {
		return nil, errors.New("wavetable: expected 20000 harmonics")
	}
	tables := []*table{}
	increment := math.Pow(2, 0.5)
	for freq := 20.0; freq < 20000.0; freq *= increment {
		harmonics := int(math.Max(1.0, 20000.0/(freq*increment)))
		samples := make([]float64, tableSize(harmonics))
		for i := 0; i < harmonics; i++ {
			sine(samples, i+1, amp[i])
		}
		normalize(samples)
		tables = append(tables, &table{
			samples:         samples,
			sizeMask:        len(samples) - 1,
			phaseTableRatio: float64(len(samples)) / twoPi,
		})
	}
	return &WaveTable{tables: tables}, nil
}

func NewSine() *WaveTable {
	amp := make([]float64, 20000)
	amp[0] = 1.0
	sine, err := New(amp)
	if err != nil {
		panic(err)
	}
	return sine
}

func NewTriangle() *WaveTable {
	amp := make([]float64, 20000)
	inverter := 1.0
	for i := 1; i <= len(amp); i++ {
		if i%2 == 1 {
			amp[i-1] = inverter * 1 / float64(i*i)
			inverter *= -1.0
		}
	}
	triangle, err := New(amp)
	if err != nil {
		panic(err)
	}
	return triangle
}

func NewSquare() *WaveTable {
	amp := make([]float64, 20000)
	for i := 1; i <= len(amp); i++ {
		if i%2 == 1 {
			amp[i-1] = 1 / float64(i)
		}
	}
	square, err := New(amp)
	if err != nil {
		panic(err)
	}
	return square
}

func NewSaw() *WaveTable {
	amp := make([]float64, 20000)
	for i := 1; i <= len(amp); i++ {
		amp[i-1] = 1 / float64(i)
	}
	saw, err := New(amp)
	if err != nil {
		panic(err)
	}
	return saw
}

const (
	oversample   = 2.0
	minTableSize = 256 // must be a power of 2
)

func tableSize(harmonics int) int {
	nyquist := float64(harmonics * 2)
	minSize := math.Max(minTableSize, nyquist*oversample)
	return 1 << uint(math.Ceil(math.Log2(minSize)))
}

func sine(samples []float64, harmonic int, amp float64) {
	indexPhaseRatio := float64(harmonic) * twoPi / float64(len(samples))
	for i := 0; i < len(samples); i++ {
		samples[i] += amp * math.Sin(float64(i)*indexPhaseRatio)
	}
}

func normalize(samples []float64) {
	m := 0.0
	for _, s := range samples {
		m = math.Max(m, math.Abs(s))
	}
	if m == 0.0 {
		return
	}
	scale := 1.0 / m
	for i, s := range samples {
		samples[i] = s * scale
	}
}

func (w *WaveTable) Add(out []float64, phase, freq float64, rate int) float64 {
	dphase := freq * twoPi / float64(rate)
	for i := range out {
		out[i] += w.Lerp(phase, freq)
		phase += dphase
		if phase > twoPi {
			phase -= twoPi
		}
	}
	return phase
}

func (w *WaveTable) Lerp(phase, freq float64) float64 {
	t := w.tables[int(math.Max(0.0, math.Log2(freq/20)*2))]
	findex := phase * t.phaseTableRatio
	index := int(findex)
	left := t.samples[index]
	right := t.samples[(index+1)&t.sizeMask]
	return left + (right-left)*(findex-float64(index))
}
