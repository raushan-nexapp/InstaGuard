package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// httpClient is shared across all commands.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// apiGet performs a GET and decodes the JSON body into dst.
func apiGet(path string, dst any) error {
	resp, err := httpClient.Get(apiURL + path)
	if err != nil {
		return fmt.Errorf("connect to %s: %w (is the engine running?)", apiURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: %s", resp.Status, string(body))
	}
	if dst == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

// apiSend performs POST/PUT/DELETE with an optional JSON body and decodes the response.
func apiSend(method, path string, body any, dst any) error {
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		buf = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, apiURL+path, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connect to %s: %w (is the engine running?)", apiURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: %s", resp.Status, string(b))
	}
	if dst == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}
