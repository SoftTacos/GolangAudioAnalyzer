package main

import (
	"AudioServer/analyzers"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/gorilla/websocket"
)

//these are sitting here while I get details figured out
var visualizerPageHTML []byte
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var am AudioManager
var fftID uint8

//TODO: Will need to account for varying sampling rates in songs, use beep.Resample
type AudioManager struct {
	streamers      []beep.Streamer
	analyzers      map[uint8]analyzers.Analyzer //all streams become a single audio stream, but each analyzer has different outputs and effects, need to be accessed individually
	format         beep.Format
	nextAnalyzerID uint8
}

//should be able to pass in a stream(buffered?) and analyzer/analyzer type and have the manager create the streamer and link them with a channel
//returns the ID of the analyzer for ez access
//TODO: this method has too much different functionality in it
func (am *AudioManager) Pair(buffer beep.StreamSeeker, analyzer analyzers.Analyzer) uint8 {
	sampleChan := make(chan [][2]float64, 1)
	listener := analyzers.Listener{
		Streamer: buffer,
		Samples:  sampleChan,
	}

	am.AddStreamer(listener)
	fmt.Println(&sampleChan)
	//analyzer.SetInputChannel(&sampleChan)

	bob := analyzer.(analyzers.FFTAnalyzer)
	bob.Samples = sampleChan

	speaker.Play(listener)
	return am.AddAnalyzer(bob)
}

/*
func (am *AudioManager) AddPair(listener *analyzers.Listener, analyzer analyzers.Analyzer) uint8 {
	am.AddStreamer(listener)
	analyzer.SetInputChannel(listener.Samples)
	return am.AddAnalyzer(analyzer)
}
*/
func (am *AudioManager) AddStreamer(streamer beep.Streamer) {
	am.streamers = append(am.streamers, streamer)
}

func (am *AudioManager) AddAnalyzer(streamer analyzers.Analyzer) uint8 {
	am.analyzers[am.nextAnalyzerID] = streamer
	am.nextAnalyzerID++
	return am.nextAnalyzerID - 1
}

func (am *AudioManager) GetAnalyzer(ID uint8) analyzers.Analyzer {
	return am.analyzers[ID]
}

func loadAudioBufferedStream(filename string) beep.StreamSeeker {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	//5120 = 512*10. Making every sample a power of 2 makes the FFT work better/faster

	fmt.Println("LOADING: ", filename)
	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)
	streamer.Close()
	fmt.Println("LOADED: ", filename)
	bs := buffer.Streamer(0, buffer.Len())
	return bs
}

func (am *AudioManager) LoadAudio(filename string) {

}

func visualizerPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, string(visualizerPageHTML))
}

func blankPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This page left blank")
}

//if they hit this page, that means they are requesting the socket
func visualizerSocketSetup(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Vis Socket Connected!")
	visualizerSocketStream(ws)
}

//this only sends, for now we don't need to listen to the socket
//messageType is an int and can be 1:Text([]uint8|[]byte), 2:binary(), 8:closemessage, 9:ping message, 10:pong message?
func visualizerSocketStream(socket *websocket.Conn) {
	ffta := am.GetAnalyzer(fftID).(*analyzers.FFTAnalyzer)
	frequencies := <-ffta.Frequencies
	for open := true; open; {
		if err := socket.WriteMessage(2, F64StoBS(frequencies)); err != nil {
			log.Println(err)
			return
		}
		frequencies = <-ffta.Frequencies
	}
}

func F64StoBS(stream []float64) []byte { //Float64 slice -> Byte slice
	byteStream := make([]byte, len(stream)*8, len(stream)*8) //*8 because a byte is uint8, 8*8=64
	for i, float := range stream {                           //float64bits takes in our float and returns IEEE 754 binary representation
		binary.LittleEndian.PutUint64(byteStream[i*8:(i+1)*8], math.Float64bits(float)) //putUint64 takes the individual float64 as binary and converts it to a uint slice in the little endian format
	}
	return byteStream
}

//static init function for now, want to get waveform streaming first
func setRoutes() {
	http.HandleFunc("/", blankPage)
	http.HandleFunc("/visualizer", visualizerPage)
	http.HandleFunc("/visualizerSocket", visualizerSocketSetup)
}

func LoadTextFile(filename string) []byte {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	filestring, err := ioutil.ReadAll(f)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	err = f.Close()
	if err != nil {
		panic(err)
	}
	return filestring
}

func wrapWithListener(buffer *beep.StreamSeeker) (*analyzers.Listener, chan [][2]float64) {
	sampleChan := make(chan [][2]float64, 1)
	listener := &analyzers.Listener{
		Streamer: *buffer,
		Samples:  sampleChan,
	}
	return listener, sampleChan
}

func main() {
	speakerFormat := beep.Format{
		SampleRate:  44100,
		NumChannels: 2,
		Precision:   4,
	}
	am = AudioManager{
		format:         speakerFormat,
		analyzers:      make(map[uint8]analyzers.Analyzer),
		nextAnalyzerID: 0,
	}
	speaker.Init(am.format.SampleRate, 5120) //samples per second, number of samples to store in the buffer
	filename := "TestSong.mp3"
	bufferedStream := loadAudioBufferedStream(filename)
	fft := &analyzers.FFTAnalyzer{
		Frequencies: make(chan []float64, 1),
	}
	//fmt.Println("PEAR")
	//fftID = am.Pair(bufferedStream, fft)
	//fmt.Println("PAIR")

	//@@@@@@@@@@@@@@@@@@@

	sampleChan := make(chan [][2]float64, 1)
	listener := analyzers.Listener{
		Streamer: bufferedStream,
		Samples:  sampleChan,
	}

	am.AddStreamer(listener)
	fmt.Println(&sampleChan)
	//analyzer.SetInputChannel(&sampleChan)

	//bob := analyzer.(analyzers.FFTAnalyzer)
	//bob.Samples = sampleChan
	fft.Samples = sampleChan
	//fft.SetInputChannel(&sampleChan)
	//f := analyzers.FFTAnalyzer.SetInputChannel
	//analyzers.FFTAnalyzer.SetInputChannel(*fft, &sampleChan)
	speaker.Play(listener)
	am.AddAnalyzer(fft) //ID :=

	//@@@@@@@@@@@@@@@@@@@

	fft.Start() //this creates the analyzer's own goroutine

	//TODO: refactor code to access AM to get to the fft

	visualizerPageHTML = LoadTextFile("visualizer.html")
	setRoutes()
	log.Fatal(http.ListenAndServe(":8082", nil))
}
