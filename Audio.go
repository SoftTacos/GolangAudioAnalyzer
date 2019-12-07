package main

//file is littered with notes while I get things sorted out properly. Goal of this is to make one type of audio analyzer function to get one solid use-case slyce. Then build a package-like interface out of it for practice

//speaker holds a mixer, play just tosses the stream into the mixer's playlist(heh)
//mixer is a collection of streamers, it's Stream() gets called
//a buffer is a streamer(?), and it can hold multiple streamers
//
import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type Listener struct {
	Streamer beep.Streamer
	samples  chan<- [][2]float64
}

//https://github.com/faiface/beep/wiki/Composing-and-controlling
//https://gobyexample.com/closures
//https://tour.golang.org/moretypes/25

/*for i := range samples {
	samples[i][0] =
	samples[i][1] =
}*/
func (l Listener) Stream(samples [][2]float64) (int, bool) { //int n, bool ok
	if l.Streamer == nil {
		return 0, false
	}

	return l.Streamer.Stream(samples)
}

func (l Listener) Err() error {
	if l.Streamer == nil {
		return nil
	}
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

type Analyzer interface {
	Start()
	Stop()
	Sampler()
}

type BeatDetector struct {
	samples <-chan [][2]float64
	stop    bool
}

func (bd BeatDetector) Sampler() {

}

func (bd BeatDetector) Start() {
	go func() {
		for !bd.stop {
			bd.Sampler()
		}
	}()

}

func (bd BeatDetector) Stop() {
	bd.stop = true
}

func main() {
	f, err := os.Open("Ben_Dust_-_Homeless_Sebastian_Groth_Power_Edit.mp3") //("Trym - BDSM.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)) //samples per second, number of samples to store in the buffer
	sampleChan := make(chan [][2]float64, 10)
	songEffects := &Listener{Streamer: streamer, samples: sampleChan}
	done := make(chan bool)

	speaker.Play(beep.Seq(songEffects, beep.Callback(func() {
		fmt.Println("DONE")
		done <- true
	})))

	bd := BeatDetector{sampleChan}
	go fmt.Println("ASDF")
	<-done
}

/*
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	buffer := beep.NewBuffer(format)
	fmt.Println("LOADING")
	buffer.Append(streamer) //this is where we actually load the file in to memory
	fmt.Println("LOADED")
	streamer.Close()

	bufferStreamSeeker := buffer.Streamer(0, buffer.Len())
	fmt.Println(reflect.TypeOf(bufferStreamSeeker))
	speaker.Play(bufferStreamSeeker)

func main() {
	filename := "Trym - BDSM.mp3"
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	//beep.StreamSeeker,, can skip around the file. len()int, Position() int, Seek(p int) error
	streamer, format, err := mp3.Decode(f) //doesn't hold the file in memory. Streamer is the....audio stream, format is the....format, like mp3, has the sample rate, etc
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()
	//sample rate, buffer size. N = number of samples, D = duration of samples
	rate := format.SampleRate * 2
	speaker.Init(rate, format.SampleRate.N(time.Second/10))
	done := make(chan bool)

	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))
	format.SampleRate.D(streamer.Position()).Round(time.Second)
	<-done

}
*/
