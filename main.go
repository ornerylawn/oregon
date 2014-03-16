package main

import (
	"code.google.com/p/portaudio-go/portaudio"
	"fmt"
	"github.com/rynlbrwn/oregon/knob"
	"github.com/rynlbrwn/oregon/polysynth"
	"math"
	"os"
	"time"
)

const (
	rate            = 44100
	channels        = 1
	framesPerBuffer = 2048
	int16Max        = 1<<15 - 1
)

var ps = polysynth.NewPolySynth()

func main() {
	ps.Volume = 0.2
	vco1 := ps.VCOs[0]
	vco1.Wave = polysynth.Saw
	vco1.Octave = 0
	vco1.Semitone = 0
	vco1.Cents = 0
	vco2 := ps.VCOs[1]
	vco2.Wave = polysynth.Saw
	vco2.Octave = 0
	vco2.Semitone = 0
	vco2.Cents = 10
	vco3 := ps.VCOs[2]
	vco3.Wave = polysynth.Saw
	vco3.Octave = -1
	vco3.Semitone = 0
	vco3.Cents = 0
	ps.AddVoice(110.0)
	ps.AddVoice(440.0 * math.Exp2(7.0/12.0))
	ps.AddVoice(880.0)

	err := knob.PrintKnobs(ps)
	if err != nil {
		fmt.Println(err)
	}

	err = portaudio.Initialize()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer portaudio.Terminate()
	stream, err := portaudio.OpenDefaultStream(0, channels, rate, framesPerBuffer, audioCallback)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer stream.Close()
	err = stream.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer stream.Stop()
	time.Sleep(5 * time.Second)
}

func audioCallback(out []int16) {
	scratch := make([]float64, len(out))
	ps.Add(scratch, rate)
	for i := range out {
		out[i] = int16(math.Min(1.0, math.Max(-1.0, scratch[i])) * int16Max)
	}
}
