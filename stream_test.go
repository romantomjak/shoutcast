package shoutcast

import (
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

func assertEqual(t *testing.T, got, want interface{}) bool {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
		return false
	}
	return true
}

func assertNoError(t *testing.T, err error) bool {
	t.Helper()
	if err != nil {
		t.Errorf("Received unexpected error:\n%+v", err)
		return false
	}
	return true
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if !assertNoError(t, err) {
		t.FailNow()
	}
}

func assertError(t *testing.T, err error) bool {
	t.Helper()
	if err == nil {
		t.Error("An error is expected but got nil.")
		return false
	}
	return true
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

func TestMissingBitrate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header()["icy-metaint"] = []string{"100"}
		w.WriteHeader(200)
	}))

	_, err := Open(ts.URL)
	assertNoError(t, err)
}

func TestUnexpectedEOF(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("icy-br", "192")
		w.Header().Set("icy-metaint", "1")

		metadata := makeMetadata("SongTitle='Prospa Prayer';")
		stream := insertMetadata([]byte{1, 1}, metadata, 1)
		// fmt.Printf("%v\n", stream)
		// unexpected EOF in the middle of a metadata block
		w.Write(stream[:len(stream)-10])
	}))
	defer ts.Close()

	s, _ := Open(ts.URL)

	b1 := make([]byte, 1)
	n, err := s.Read(b1)
	if assertNoError(t, err) && assertEqual(t, 1, n) {
		assertEqual(t, []byte{1}, b1)
	}

	// The metadata is immediatly read and does not fit into the buffer.
	// -> `0, nil` is returned.
	// Filling the buffer after the reading of the metadata would be more complexity without advantage.
	n, err = s.Read(b1)
	assertNoError(t, err)
	assertEqual(t, 0, n)

	b2 := make([]byte, 1)
	n, err = s.Read(b2)
	if assertNoError(t, err) && assertEqual(t, 1, n) {
		assertEqual(t, []byte{1}, b2)
	}

	// ooops, nothing to read
	b3 := make([]byte, 1)
	n, err = s.Read(b3)
	assertEqual(t, 0, n)
	if assertError(t, err) {
		assertEqual(t, io.ErrUnexpectedEOF, err)
	}
}

func TestMetaintEqualsClientBufferLength(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("icy-br", "192")
		w.Header().Set("icy-metaint", "2")

		metadata := makeMetadata("SongTitle='Prospa Prayer';")
		stream := insertMetadata([]byte{1, 1, 1, 1, 1, 1}, metadata, 2)
		// fmt.Printf("%v\n", stream)
		w.Write(stream)
	}))
	defer ts.Close()

	s, _ := Open(ts.URL)

	b1 := make([]byte, 2)
	n, err := s.Read(b1)
	if assertNoError(t, err) && assertEqual(t, 2, n) {
		assertEqual(t, []byte{1, 1}, b1)
	}

	// The metadata is immediatly read and does not fit into the buffer.
	// -> `0, nil` is returned.
	// Filling the buffer after the reading of the metadata would be more complexity without advantage.
	n, err = s.Read(b1)
	assertNoError(t, err)
	assertEqual(t, 0, n)

	b2 := make([]byte, 2)
	n, err = s.Read(b2)
	if assertNoError(t, err) && assertEqual(t, 2, n) {
		assertEqual(t, []byte{1, 1}, b2)
	}

	// no data except metadata read, again
	n, err = s.Read(b2)
	assertNoError(t, err)
	assertEqual(t, 0, n)

	b3 := make([]byte, 2)
	n, err = s.Read(b3)
	if assertNoError(t, err) && assertEqual(t, 2, n) {
		assertEqual(t, []byte{1, 1}, b3)
	}

	// check for EOF
	n, err = s.Read(b1)
	assertEqual(t, 0, n)
	if assertError(t, err) {
		assertEqual(t, io.EOF, err)
	}
}

func TestMetaintGreaterThanClientBufferLength(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("icy-br", "192")
		w.Header().Set("icy-metaint", "3")

		metadata := makeMetadata("SongTitle='Prospa Prayer';")
		stream := insertMetadata([]byte{1, 1, 1, 1, 1, 1}, metadata, 3)
		// fmt.Printf("%v\n", stream)
		w.Write(stream)
	}))
	defer ts.Close()

	s, _ := Open(ts.URL)

	b1 := make([]byte, 2)
	n, err := s.Read(b1)
	if assertNoError(t, err) && assertEqual(t, 2, n) {
		assertEqual(t, []byte{1, 1}, b1)
	}

	// only one byte read then follows metadata
	b2 := make([]byte, 2)
	n, err = s.Read(b2)
	if assertNoError(t, err) && assertEqual(t, 1, n) {
		// don't assert the second byte, only one read
		assertEqual(t, []byte{1}, b2[:1])
	}

	b3 := make([]byte, 2)
	n, err = s.Read(b3)
	if assertNoError(t, err) && assertEqual(t, 2, n) {
		assertEqual(t, []byte{1, 1}, b3)
	}

	// only one byte read then follows metadata and then EOF
	b4 := make([]byte, 2)
	n, err = s.Read(b4)
	if assertEqual(t, 1, n) {
		// don't assert the second byte, only one read
		assertEqual(t, []byte{1}, b4[:1])
	}
	if assertError(t, err) {
		assertEqual(t, io.EOF, err)
	}
}

func TestClientBufferLargeEnoughForMetadata(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("icy-br", "192")
		w.Header().Set("icy-metaint", "3")

		metadata := makeMetadata("SongTitle='Prospa Prayer';")
		stream := insertMetadata([]byte{3, 4, 5, 6, 7, 8, 9}, metadata, 3)
		w.Write(stream)
	}))
	defer ts.Close()

	s, err := Open(ts.URL)
	requireNoError(t, err)

	// metadata length is 33 (2*16+1) -> 38 - 33 = 5 bytes stream data to be read
	b1 := make([]byte, 38)
	n, err := s.Read(b1)
	if assertNoError(t, err) && assertEqual(t, 5, n) {
		assertEqual(t, []byte{3, 4, 5, 6, 7}, b1[:5])
	}

	b2 := make([]byte, 38)
	n, err = s.Read(b2)
	if assertEqual(t, 2, n) {
		assertEqual(t, []byte{8, 9}, b2[:2])
	}
	if assertError(t, err) {
		assertEqual(t, io.EOF, err)
	}
}

func TestClientBufferLargeEnoughForTwoTimesMetadata(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("icy-br", "192")
		w.Header().Set("icy-metaint", "3")

		metadata := makeMetadata("SongTitle='Prospa Prayer';")
		stream := insertMetadata([]byte{3, 4, 5, 6, 7, 8, 9, 10}, metadata, 3)
		w.Write(stream)
	}))
	defer ts.Close()

	s, err := Open(ts.URL)
	requireNoError(t, err)

	// metadata length is 33 (2*16+1) -> 73 - 2 * 33 = 7
	b1 := make([]byte, 73)
	n, err := s.Read(b1)
	if assertNoError(t, err) && assertEqual(t, 7, n) {
		assertEqual(t, []byte{3, 4, 5, 6, 7, 8, 9}, b1[:7])
	}

	b2 := make([]byte, 38)
	n, err = s.Read(b2)
	if assertEqual(t, 1, n) {
		assertEqual(t, []byte{10}, b2[:1])
	}
	if assertError(t, err) {
		assertEqual(t, io.EOF, err)
	}
}

// test for EOF
// test for read on closed socket
