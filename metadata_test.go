package shoutcast

import "testing"

func TestCanParseMetadata(t *testing.T) {
	buf := []byte("StreamTitle='Prospa - Prayer';")
	m := NewMetadata(buf)

	assertStrings(t, m.StreamTitle, "Prospa - Prayer")
}

func TestCanCompareTwoMetadataStructs(t *testing.T) {
	buf := []byte("StreamTitle='Prospa - Prayer';")
	m := NewMetadata(buf)
	other := NewMetadata(buf)

	assertEqual(t, m.Equals(other), true)
}

func TestCanCompareNilStructs(t *testing.T) {
	buf := []byte("StreamTitle='Prospa - Prayer';")
	m := NewMetadata(buf)

	assertEqual(t, m.Equals(nil), false)
}
