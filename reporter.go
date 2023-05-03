package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/speps/go-hashids/v2"
)

func sendTestRequest(reporterid int, payload Match) {
	reporter := make([]int, 0)
	reporter = append(reporter, reporterid)

	hd := hashids.NewData()
	hd.Salt = HashSaltParam
	hd.MinLength = 64
	hd.Alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode(reporter)

	url := "http://127.0.0.1:3000/"
	contentType := "application/json"
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("X-Reporter", e)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// fmt.Println(string(body))
}
