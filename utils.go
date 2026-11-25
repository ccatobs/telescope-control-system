package main

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"io"
	"net/http"
	"time"
)

// postJSONClient has a timeout, since postJSON is called from program track
// uploads, which the main loop waits for.
var postJSONClient = &http.Client{Timeout: 2 * time.Second}

// writeJSON writes v to w as JSON, followed by a newline.
func writeJSON(w io.Writer, v any) error {
	return json.MarshalEncode(jsontext.NewEncoder(w), v)
}

func postJSON(url string, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	resp, err := postJSONClient.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return err
}
