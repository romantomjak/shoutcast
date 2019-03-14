# shoutcast

Go interface for SHOUTcast streaming protocol. 

---

Example SHOUTcast client:

```go
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
    defer stream.Close()

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
```

## Shoutcast protocol

Very little information is available about how the streaming technology
actually works. Everything in this repo was pieced together from random
sites and [Wayback Machine](https://archive.org/web/) archives.

The protocol itself is built on top of HTTP. The client sends information
about itself in the form of HTTP headers, and if it can handle title streaming
it also sends and extra header:

```
icy-metadata: 1
```

This header signifies that the client has the ability to understand the title
streaming tags coming from the server. When set, the server will embed metadata
about the stream at periodic intervals (once every `icy-metaint` bytes) in the
encoded audio stream itself.

The value of `icy-metaint` is decided by the shoutcast server configuration and
is sent to the client as part of the initial reply.

The first byte of that metadata block tells you the length of the data.
A length of 0 means there is no updated metadata.

- https://www.radiotoolbox.com/community/forums/viewtopic.php?t=74
- https://stackoverflow.com/q/12407248/758384

## License

MIT
