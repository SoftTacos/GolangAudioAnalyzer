package analyzers

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/faiface/beep"
)

//Listener is a wrapper for a stream, it just tosses the passed in samples to a channel. It is a streamer
type Listener struct {
	Streamer beep.Streamer
	Samples  chan<- [][2]float64
	//might want to have a buffer?
}

func (l Listener) Stream(samples [][2]float64) (int, bool) { //int n, bool ok
	if l.Streamer == nil {
		fmt.Println("ERROR")
		return 0, false
	}
	//fmt.Println("STREAMING", len(samples))
	//select {
	//case l.Samples <- samples:
	l.Samples <- samples //TODO: if l.Samples is full, dump
	//}
	return l.Streamer.Stream(samples)
}

func (l Listener) Err() error {
	if l.Streamer == nil {
		fmt.Println("Error: Streamer is nil")
		return nil
	}
	fmt.Println("ERR ERROR")
	return l.Streamer.Err()
}

type Analyzer interface {
	SetInputChannel(channel *chan [][2]float64)
	Stream()
	//GetData() AnalyzerData
}

type AnalyzerData struct {
	Type   string
	floats []float64
}

func (ad *AnalyzerData) Data() {

}

type FFTAnalyzer struct {
	Samples     <-chan [][2]float64
	Frequencies chan []float64
	Stopped     bool
}

func (ffta FFTAnalyzer) SetInputChannel(channel *chan [][2]float64) {
	ffta.Samples = *channel
}

func (ffta FFTAnalyzer) Stream() {
	//TODO: allocate variables here before loop
	//TODO: refactor select statement to wait the thread if samples is empty
	select {
	case samples := <-ffta.Samples:
		//TBD if FFTAnalyzer should check if len(samples) is a power of 2. Keeps the math easy
		//power := math.Log2(float64(len(samples)))
		cs := make([]complex128, len(samples))
		samplesFFTch1 := make([]float64, len(samples))
		for i := range samples {
			samplesFFTch1[i] = samples[i][0]
		}
		FFT(samplesFFTch1, cs, len(samplesFFTch1), 1)
		ffta.Frequencies <- samplesFFTch1
	default:
		//fmt.Println("SAMPLES IS EMPTY: ", &samples)
	}
}

func (ffta *FFTAnalyzer) Start() {
	go func() {
		for !ffta.Stopped {
			ffta.Stream()
		}
	}()

}

func (ffta *FFTAnalyzer) Stop() {
	ffta.Stopped = true
}

//TODO
func (ffta FFTAnalyzer) Err() error {
	return nil
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

type LowPassFilter struct {
	Samples     <-chan [][2]float64
	Frequencies chan []float64
	stopped     bool
}

func (lpf *LowPassFilter) Stream() {

}

func (lpf *LowPassFilter) Start() {
	go func() {
		for !lpf.stopped {
			lpf.Stream()
		}
	}()

}

func (lpf *LowPassFilter) Stop() {
	lpf.stopped = true
}
