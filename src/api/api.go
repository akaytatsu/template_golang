package api

import (
	"log"
	"net/http"

	"app/api/handlers"
	"app/api/middleware"
	"app/config"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

func StartWebServer() {
	config.ReadEnvironmentVars()

	r := mux.NewRouter()

	n := negroni.New(
		negroni.HandlerFunc(middleware.Cors),
		negroni.NewLogger(),
		negroni.NewRecovery(),
	)

	handlers.MakeApiHandlers(r, *n)

	n.UseHandler(r)

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8080", n))
}
