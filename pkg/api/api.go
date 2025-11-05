package api

import "net/http"

const DATEFORMAT = "20060102"

func Init() {
	http.HandleFunc("/api/nextdate", nextDateHandler)
}
