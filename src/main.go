package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Golang is running!\n"))
}

func main() {

	r := mux.NewRouter()

	// Routes consist of a path and a handler function.
	r.HandleFunc("/", RootHandler)

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8080", r))
}
