package main

//file is littered with notes while I get things sorted out properly. Goal of this is to make one type of audio analyzer function to get one solid use-case slyce. Then build a package-like interface out of it for practice
//TestSong.mp3 is ignored because copyrighting

//NOTES:
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

//Listener is a wrapper for a streamer, it just tosses the passed in samples to a channel
type Listener struct {
	Streamer beep.Streamer
	samples  chan<- [][2]float64
	//might want to have a buffer?
}

func (l Listener) Stream(samples [][2]float64) (int, bool) { //int n, bool ok
	if l.Streamer == nil {
		return 0, false
	}
	l.samples <- samples
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

//an analyzer is made to then take those samples passed through the sample channel and...analyzes them!
type Analyzer interface {
	Start()
	Stop()
	Sampler()
}

type TestAnalyzer struct {
	samples <-chan [][2]float64
	stop    bool
}

func (ta TestAnalyzer) Sampler() {
	temp := <-ta.samples
}

func (ta TestAnalyzer) Start() {
	//fmt.Println("Starting the BD")
	go func() {
		for ta.stop {
			ta.Sampler()
		}
	}()

}

func (ta TestAnalyzer) Stop() {
	//fmt.Println("Stopping the BD")
	ta.stop = true
}

func main() {
	f, err := os.Open("TestSong.mp3")
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

	//done := make(chan bool)
	speaker.Play(beep.Seq(songEffects, beep.Callback(func() {
		fmt.Println("DONE")
		//done <- true
	})))

	//
	ta := TestAnalyzer{samples: sampleChan, stop: true}
	ta.Start()
	time.Sleep(5 * time.Second)
	ta.Stop()
	time.Sleep(2 * time.Second)

	//<-done
}
