package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/speps/go-hashids/v2"
)

func sendMatchReport(reporterid int, payload Match) {
	log.Printf("Start to try to send report")
	reporter := make([]int, 0)
	reporter = append(reporter, reporterid)

	hd := hashids.NewData()
	hd.Salt = HashSaltParam
	hd.MinLength = 64
	hd.Alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode(reporter)

	url := fmt.Sprintf("%s/v1/submitreport", ReportServer)
	log.Printf("Report URL: %s", url)
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
	req.Header.Add("X-Reporter", e)

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	log.Printf("End to try to send report")
}
