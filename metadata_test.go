package shoutcast

import "testing"

func TestCanParseMetadata(t *testing.T) {
	buf := []byte("StreamTitle='Prospa - Prayer';")
	m := NewMetadata(buf)

	assertStrings(t, m.StreamTitle, "Prospa - Prayer")
}
