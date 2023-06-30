package main

import (
	"fmt"
	"io"

	mp3 "github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"

	"github.com/romantomjak/shoutcast"
)

// shoutcast URLs change, you might need to change this to a URL that is live
const ShoutcastUrl = "https://dancewave.online:443/dance.mp3"

func main() {
	fmt.Println("Processing URL:", ShoutcastUrl)

	// open stream
	stream, err := shoutcast.Open(ShoutcastUrl)
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
	defer stream.Close()

	// initialise audio player
	playerCtx, err := oto.NewContext(decoder.SampleRate(), 2, 2, 8192)
	if err != nil {
		panic(err)
	}

	player := playerCtx.NewPlayer()
	defer player.Close()

	// enjoy the music
	if _, err := io.Copy(player, decoder); err != nil {
		panic(err)
	}
}
