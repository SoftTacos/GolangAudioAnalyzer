package analyzers

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/faiface/beep"
)

//Listener is a wrapper for a streamer, it just tosses the passed in samples to a channel
type Listener struct {
	Streamer beep.Streamer
	Samples  chan<- [][2]float64
	//might want to have a buffer?
}

func (l *Listener) Stream(samples [][2]float64) (int, bool) { //int n, bool ok
	if l.Streamer == nil {
		fmt.Println("ERROR")
		return 0, false
	}
	//fmt.Println("STREAMING", len(samples))
	l.Samples <- samples
	return l.Streamer.Stream(samples)
}

func (l *Listener) Err() error {
	if l.Streamer == nil {
		return nil
	}
	fmt.Println("ERR ERROR")
	return l.Streamer.Err()
}

//I have no IRL audio experiene so idk what this is analogous to, will rename later
type ThisHoldsLotsOfAnalyzers struct {
	analyzers []Analyzer
}

func (thloa ThisHoldsLotsOfAnalyzers) Start() {
	for _, a := range thloa.analyzers {
		a.Start()
	}
}

//an analyzer is made to then take those samples passed through the sample channel and...analyzes them!
type Analyzer interface {
	Start()
	Stop()
	Sampler()
}

type FFTAnalyzer struct {
	Samples     <-chan [][2]float64
	Frequencies chan []float64
	Stopped     bool
}

func (ta *FFTAnalyzer) Sampler() {
	select {
	case Samples := <-ta.Samples:
		cs := make([]complex128, len(Samples))
		samplesFFTch1 := make([]float64, len(Samples))
		for i := range Samples {
			samplesFFTch1[i] = Samples[i][0]
		}
		FFT(samplesFFTch1, cs, len(samplesFFTch1), 1)
	default:

	}
}

func (ta *FFTAnalyzer) Start() {
	go func() {
		for !ta.Stopped {
			ta.Sampler()
		}
	}()

}

func (ta *FFTAnalyzer) Stop() {
	ta.Stopped = true
}

func FFT(x []float64, y []complex128, n int, s int) { //https://rosettacode.org/wiki/Fast_Fourier_transform#Go
	if n == 1 {
		y[0] = complex(x[0], 0)
		return
	}
	FFT(x, y, n/2, 2*s)
	FFT(x[s:], y[n/2:], n/2, 2*s)
	for k := 0; k < n/2; k++ {
		tf := cmplx.Rect(1, -2*math.Pi*float64(k)/float64(n)) * y[k+n/2]
		y[k], y[k+n/2] = y[k]+tf, y[k]-tf
	}
}
