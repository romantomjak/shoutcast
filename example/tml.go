package main

import (
	"io"

	mp3 "github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"

	"github.com/romantomjak/shoutcast"
)

func main() {
	// open stream
	stream, err := shoutcast.Open("http://streamingp.shoutcast.com/TomorrowlandOneWorldRadio")
	if err != nil {
		panic(err)
	}

	// optionally register a callback function to be called when song changes
	stream.MetadataCallbackFunc = func(m *shoutcast.Metadata) {
		println("Now listening to: ", m.StreamTitle)
	}

	// setup mp3 decoder
	decoder, err := mp3.NewDecoder(stream)
	if err != nil {
		panic(err)
	}
	defer decoder.Close()

	// initialise audio player
	player, err := oto.NewPlayer(decoder.SampleRate(), 2, 2, 8192)
	if err != nil {
		panic(err)
	}
	defer player.Close()

	// enjoy the music
	if _, err := io.Copy(player, decoder); err != nil {
		panic(err)
	}
}
