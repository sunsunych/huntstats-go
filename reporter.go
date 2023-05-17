package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/speps/go-hashids/v2"
)

type ApiResponseMessage struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Path    string `json:"path"`
}

func sendMatchReport(reporterid int, payload Match) {
	reporterKey := buildReporterHash(reporterid)
	log.Printf("Sending match report")
	url := fmt.Sprintf("%s/v1/submitreport", ReportServer)
	contentType := "application/json"
	data, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("X-Reporter", reporterKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	var respBodyJson = ApiResponseMessage{}

	respBodyByte, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(respBodyByte, &respBodyJson)
	log.Printf("Response Code: %d - %s", resp.StatusCode, respBodyJson.Message)
	defer resp.Body.Close()
}

func buildReporterHash(reporterid int) string {
	reporter := make([]int, 0)
	reporter = append(reporter, reporterid)

	hd := hashids.NewData()
	hd.Salt = HashSaltParam
	hd.MinLength = 64
	hd.Alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode(reporter)
	return e
}

func buildReporterProfileLink(reporterid int) string {
	authkey := buildReporterHash(reporterid)
	url := fmt.Sprintf("%s/login?authkey=%s", ProfileServer, authkey)
	return url
}
