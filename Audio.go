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
	"time"

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
var ta analyzers.FFTAnalyzer

func visualizerPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, string(visualizerPageHTML)) //lazy for now
}

func blankPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This page left blank")
}

//if they hit this page, that means they are requesting the socket for simplicity's sake
func visualizerSocketSetup(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade this connection to a WebSocket
	// connection
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
	//test send
	var frequencies []float64
	for open := true; open; {
		frequencies = <-ta.Frequencies
		if err := socket.WriteMessage(2, F64StoBS(frequencies)); err != nil {
			log.Println(err)
			return
		}
	}
	frequencies = []float64{} //really go
}

func F64StoBS(stream []float64) []byte { //Float64 slice -> Byte slice
	byteStream := make([]byte, len(stream)*8, len(stream)*8) //*8 because a byte is uint8, 8*8=64
	for i, float := range stream {
		binary.LittleEndian.PutUint64(byteStream[i*8:(i+1)*8], math.Float64bits(float))
	}
	return byteStream
}

func FStouI8S(stream []float64) []uint8 { //Float64 slice -> int8 slice

	return []uint8{0, 0, 0, 0}
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

//placeholder name
func audioStart() {
	filename := "TestSong2.mp3"
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	//defer streamer.Close()
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)) //samples per second, number of samples to store in the buffer
	fmt.Println("LOADING SONG")
	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)
	streamer.Close()
	fmt.Println("LOADED")
	bs := buffer.Streamer(0, buffer.Len())

	sampleChan := make(chan [][2]float64, 10)
	songListener := &analyzers.Listener{
		Streamer: bs,
		Samples:  sampleChan,
	}

	speaker.Play(beep.Seq(songListener, beep.Callback(func() {
		fmt.Println("Playing: ")
	})))
	ta = analyzers.FFTAnalyzer{}
	ta.Samples = sampleChan
	freqChan := make(chan []float64, 10)
	ta.Frequencies = freqChan

	ta.Start()
	//STUFF
	time.Sleep(100 * time.Second)
	fmt.Println("Shutting down audio")
	ta.Stop()
}

func main() {
	go audioStart() //placeholder for now, will eventually use this to start the audio analysis loop

	visualizerPageHTML = LoadTextFile("visualizer.html")
	setRoutes()
	log.Fatal(http.ListenAndServe(":8082", nil))
}
