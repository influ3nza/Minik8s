package network

import (
	"bytes"
	"encoding/json"
	"errors"
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

	dataStr := ""

	data := result["data"]
	if data != nil {
		dataStr = fmt.Sprint(data)
	}

	errs := result["error"]
	if errs != nil {
		err = errors.New(fmt.Sprint(errs))
	}

	return dataStr, err
}

func PostRequest(uri string, req_body []byte) (string, error) {
	body := bytes.NewReader(req_body)
	resp, err := http.Post(uri, "application/json", body)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&result)

	dataStr := ""

	data := result["data"]
	if data != nil {
		dataStr = fmt.Sprint(data)
	}

	errs := result["error"]
	if errs != nil {
		err = errors.New(fmt.Sprint(errs))
	}

	return dataStr, err
}

func DelRequest(uri string) (string, error) {
	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&result)

	dataStr := ""

	data := result["data"]
	if data != nil {
		dataStr = fmt.Sprint(data)
	}

	errs := result["error"]
	if errs != nil {
		err = errors.New(fmt.Sprint(errs))
	}

	return dataStr, err
}
