package polysynth

import (
	"fmt"
	"github.com/rynlbrwn/oregon/wavetable"
	"math"
	"time"
)

type Wave int

const (
	Sine Wave = iota
	Triangle
	Square
	Saw
)

var (
	sine     = wavetable.NewSine()
	triangle = wavetable.NewTriangle()
	square   = wavetable.NewSquare()
	saw      = wavetable.NewSaw()
)

type VCO struct {
	Wave     Wave `knob:"{sine,triangle,square,saw}"`
	Octave   int  `knob:",-2,2,1"`
	Semitone int  `knob:",-12,12,1"`
	Cents    int  `knob:",-100,100,1"`
}

func (o *VCO) Add(out []float64, phase, freq float64, rate int) float64 {
	f := freq * math.Exp2((float64(12*o.Octave+o.Semitone)+float64(o.Cents)/100)/12)
	switch o.Wave {
	case Sine:
		return sine.Add(out, phase, f, rate)
	case Triangle:
		return triangle.Add(out, phase, f, rate)
	case Square:
		return square.Add(out, phase, f, rate)
	case Saw:
		return saw.Add(out, phase, f, rate)
	default:
		panic(fmt.Sprintf("polysynth: unknown wave type %v", o.Wave))
	}
}

type ADSR struct {
	Attack  time.Duration `knob:"ms,0.0,1000.0,linear"`
	Decay   time.Duration `knob:"ms,0.0,1000.0,linear"`
	Sustain float64       `knob:",,0.0,1.0,linear"`
	Release time.Duration `knob:"ms,0.0,1000.0,linear"`
}

func NewADSR() *ADSR {
	return &ADSR{Sustain: 1.0}
}

func (a *ADSR) Mul(out []float64, t time.Duration, gate Gate, rate int, amp float64) (time.Duration, Gate) {
	for i := range out {
		out[i] *= amp
	}
	return t, gate
}

type PolySynth struct {
	VCOs []*VCO
	ADSR *ADSR
	// TODO: volume should be decibels but its an amplitude right now.
	Volume float64 `knob:"db,-40.0,0.0,logarithmic"`
	Voices []*Voice
}

type Gate int

const (
	GateOn = iota
	GateRelease
	GateOff
)

type Voice struct {
	Freq   float64
	Gate   Gate
	phases []float64
	t      time.Duration
}

const vcoCount = 3

func NewVoice(freq float64) *Voice {
	return &Voice{
		Freq:   freq,
		phases: make([]float64, vcoCount),
	}
}

func NewPolySynth() *PolySynth {
	vcos := make([]*VCO, vcoCount)
	for i := range vcos {
		vcos[i] = &VCO{}
	}
	return &PolySynth{
		VCOs:   vcos,
		ADSR:   NewADSR(),
		Voices: []*Voice{},
	}
}

func (p *PolySynth) AddVoice(freq float64) {
	p.Voices = append(p.Voices, NewVoice(freq))
}

func (p *PolySynth) Add(out []float64, rate int) {
	scratch := make([]float64, len(out))
	for _, v := range p.Voices {
		clear(scratch)
		for i := range v.phases {
			v.phases[i] = p.VCOs[i].Add(scratch, v.phases[i], v.Freq, rate)
		}
		// TODO: clean up voices whose gate is GateOff.
		v.t, v.Gate = p.ADSR.Mul(scratch, v.t, v.Gate, rate, p.Volume)
		sum(out, scratch)
	}
}

func clear(out []float64) {
	for i := range out {
		out[i] = 0.0
	}
}

func sum(dst []float64, src []float64) {
	for i := range dst {
		dst[i] += src[i]
	}
}
