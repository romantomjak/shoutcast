package main

import (
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"

	mp3 "github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"

	"github.com/romantomjak/shoutcast"
)

func parseAndValidateURL() string {
	urlStr := flag.String("url", "", "URL to process")

	// parse command-line arguments
	flag.Parse()

	// check if URL argument was provided
	if *urlStr == "" {
		fmt.Println("You must specify a URL with the -url flag.")
		printUsage()
		os.Exit(1)
	}

	// check if the URL is valid
	_, err := url.ParseRequestURI(*urlStr)
	if err != nil {
		fmt.Println("Invalid URL provided.")
		printUsage()
		os.Exit(1)
	}

	return *urlStr
}

func printUsage() {
	fmt.Println("\nUsage: go run main.go -url=<your-shoutcase-url>")
}

func main() {
	shoutcastUrl := parseAndValidateURL()

	// continue processing
	fmt.Println("Processing URL:", shoutcastUrl)

	// open stream
	stream, err := shoutcast.Open(shoutcastUrl)
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
