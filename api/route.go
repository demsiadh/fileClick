package api

import (
	"fileClick/service"
	"net/http"
)

func InitRouter() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/click", service.Click)
	mux.HandleFunc("/topN", service.GetTopN)
	mux.HandleFunc("/all", service.GetAll)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	return srv
}
