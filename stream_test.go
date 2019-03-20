package shoutcast

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func assertStrings(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func makeMetadata(s string) []byte {
	// needs to be a multiple of 16
	msize := byte(math.Ceil(float64(len(s)) / 16.0))

	buf := make([]byte, 16*msize+1, 16*msize+1)
	buf[0] = msize
	copy(buf[1:], s)

	return buf
}

// insertMetadata inserts `metadata` into `data` every `n` bytes
func insertMetadata(data []byte, metadata []byte, n int) []byte {
	numMetadata := int(math.Ceil(float64(len(data) / n)))

	bufSize := len(metadata)*numMetadata + len(data)
	buf := make([]byte, bufSize)

	for i, j := 0, 0; i < len(data); i = i + n {
		dataStart := i
		dataEnd := i + n
		if dataEnd >= len(data) {
			dataEnd = len(data)
		}

		bufStart := j * (len(metadata) + n)
		bufEnd := bufStart + (dataEnd - dataStart)

		copy(buf[bufStart:bufEnd], data[dataStart:dataEnd])
		copy(buf[bufEnd:], metadata)

		j++
	}
	return buf
}

func TestRequiredHTTPHeadersArePresent(t *testing.T) {
	var headers http.Header
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = r.Header
	}))
	defer ts.Close()

	Open(ts.URL)

	assertStrings(t, headers.Get("icy-metadata"), "1")
	assertStrings(t, headers.Get("user-agent")[:6], "iTunes")
}

func TestUnexpectedEOF(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("icy-br", "192")
		w.Header().Set("icy-metaint", "1")

		metadata := makeMetadata("SongTitle='Prospa Prayer';")
		stream := insertMetadata([]byte{1, 1}, metadata, 1)
		fmt.Printf("%v\n", stream)
		w.Write(stream)
	}))
	defer ts.Close()

	s, _ := Open(ts.URL)

	b1 := make([]byte, 1)
	s.Read(b1)
	assertEqual(t, b1, []byte{1})

	b2 := make([]byte, 1)
	s.Read(b2)
	assertEqual(t, b2, []byte{1})

	// ooops, nothing to read
	b3 := make([]byte, 1)
	_, err := s.Read(b3)
	assertEqual(t, err, io.ErrUnexpectedEOF)
}

func TestMetaintEqualsClientBufferLength(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("icy-br", "192")
		w.Header().Set("icy-metaint", "2")

		metadata := makeMetadata("SongTitle='Prospa Prayer';")
		stream := insertMetadata([]byte{1, 1, 1, 1, 1, 1}, metadata, 2)
		fmt.Printf("%v\n", stream)
		w.Write(stream)
	}))
	defer ts.Close()

	s, _ := Open(ts.URL)

	b1 := make([]byte, 2)
	s.Read(b1)
	assertEqual(t, b1, []byte{1, 1})

	b2 := make([]byte, 2)
	s.Read(b2)
	assertEqual(t, b2, []byte{1, 1})

	b3 := make([]byte, 2)
	s.Read(b3)
	assertEqual(t, b3, []byte{1, 1})
}

func TestMetaintGreaterThanClientBufferLength(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("icy-br", "192")
		w.Header().Set("icy-metaint", "3")

		metadata := makeMetadata("SongTitle='Prospa Prayer';")
		stream := insertMetadata([]byte{1, 1, 1, 1, 1, 1}, metadata, 3)
		fmt.Printf("%v\n", stream)
		w.Write(stream)
	}))
	defer ts.Close()

	s, _ := Open(ts.URL)

	b1 := make([]byte, 2)
	s.Read(b1)
	assertEqual(t, b1, []byte{1, 1})

	b2 := make([]byte, 2)
	s.Read(b2)
	assertEqual(t, b2, []byte{1, 1})

	b3 := make([]byte, 2)
	s.Read(b3)
	assertEqual(t, b3, []byte{1, 1})
}

// test for EOF
// test for read on closed socket
