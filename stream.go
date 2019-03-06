package shoutcast

import (
	"io"
	"net/http"
)

type Stream struct {
	url string
	rc  io.ReadCloser
}

func Open(url string) (*Stream, error) {
	client := http.DefaultClient

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("accept", "*/*")
	req.Header.Add("user-agent", "iTunes/12.9.2 (Macintosh; OS X 10.14.3) AppleWebKit/606.4.5")
	req.Header.Add("icy-metadata", "1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	s := &Stream{
		url: url,
		rc:  resp.Body,
	}
	return s, nil
}
