package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Info struct {
	Height int `json:"height"`
}

// this is a convenience function to get the current height of the BTC blockchain
// it's not reliable
func getBTCHeight() (int, error) {
	url := "https://blockchain.info/latestblock"
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("error making request to blockchain API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %v", err)
	}

	var info Info
	if err := json.Unmarshal(body, &info); err != nil {
		return 0, fmt.Errorf("error parsing JSON response: %v", err)
	}

	return info.Height, nil
}
