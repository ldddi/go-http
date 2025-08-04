package main

import (
	httpfileserver "httpserver/pkg"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{path:.*}", httpfileserver.BrowserGetHandler)

	http.ListenAndServe(":8888", r)
}
