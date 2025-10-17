package main

import (
	"encoding/json"

	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Device struct {
	ID int `json:"id"`
	Mac string `json:"mac"`
	Firmware string `json:"firmware"`
}

var dvs []Device

func init() {
	dvs = []Device{
		{ID: 1, Mac: "00:00:00:00:01", Firmware: "1.0.0"},
		{ID: 2, Mac: "00:00:00:00:02", Firmware: "1.0.1"},
		{ID: 3, Mac: "00:00:00:00:03", Firmware: "1.0.2"},
	}
}

func main() {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/devices", getDevices)
	http.ListenAndServe(":8081", nil)
}


func getDevices(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(dvs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}