package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func GetRequest(uri string) (string, string, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&result)

	dataStr, errStr := "", ""

	data := result["data"]
	if data != nil {
		dataStr = fmt.Sprint(data)
	}

	errs := result["error"]
	if errs != nil {
		errStr = fmt.Sprint(errs)
	}

	return dataStr, errStr, nil
}

func PostRequest(uri string, req_body []byte) (string, string, error) {
	body := bytes.NewReader(req_body)
	resp, err := http.Post(uri, "application/json", body)
	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&result)

	dataStr, errStr := "", ""

	data := result["data"]
	if data != nil {
		dataStr = fmt.Sprint(data)
	}

	errs := result["error"]
	if errs != nil {
		errStr = fmt.Sprint(errs)
	}

	return dataStr, errStr, nil
}

func DelRequest(uri string) (string, string, error) {
	//TODO: need to be tested.
	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return "", "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&result)

	dataStr, errStr := "", ""

	data := result["data"]
	if data != nil {
		dataStr = fmt.Sprint(data)
	}

	errs := result["error"]
	if errs != nil {
		errStr = fmt.Sprint(errs)
	}

	return dataStr, errStr, nil
}
