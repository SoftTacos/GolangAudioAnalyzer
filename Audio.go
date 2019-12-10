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

	"AudioServer/analyzers"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

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
	songListener := &analyzers.Listener{
		Streamer: streamer,
		Samples:  sampleChan,
	}

	speaker.Play(beep.Seq(songListener, beep.Callback(func() {
		fmt.Println("DONE")
	})))
	ta := analyzers.TestAnalyzer{}
	ta.Samples = sampleChan

	ta.Start()
	time.Sleep(500 * time.Second)
	ta.Stop()
}
