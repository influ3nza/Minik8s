package network

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func GetRequest(uri string) (string, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&result)

	data := result["data"]
	dataStr := fmt.Sprint(data)

	return dataStr, nil
}
