package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func postJSON(url string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return err
}
