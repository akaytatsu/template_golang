package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}

func MakeApiHandlers(r *mux.Router, n negroni.Negroni) {
	r.Handle("/", n.With(negroni.WrapFunc(homeHandler))).
		Methods("GET")
}
