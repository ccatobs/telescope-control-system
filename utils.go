package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

// postJSONClient has a timeout, since postJSON is called from program track
// uploads, which the main loop waits for.
var postJSONClient = &http.Client{Timeout: 2 * time.Second}

func postJSON(url string, data interface{}) error {
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
