package shoutcast

import "strings"

// Metadata represents the stream metadata sent by the server
type Metadata struct {
	StreamTitle string
}

// NewMetadata returns parsed metadata
func NewMetadata(b []byte) *Metadata {
	m := &Metadata{}

	props := strings.Split(string(b), ";")
	for _, prop := range props {
		if prop == "" {
			continue
		}
		parts := strings.Split(prop, "=")
		if parts[0] == "StreamTitle" {
			m.StreamTitle = strings.Trim(parts[1], "'")
		}
	}

	return m
}
