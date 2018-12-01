package http

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Server defines the HTTP server serving the app.
type Server struct {
	port   int
	router *mux.Router
}

// Run starts the HTTP server.
func (s *Server) Run() {
	handler := handlers.LoggingHandler(os.Stdout, s.router)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), handler); err != nil {
		panic(errors.Wrap(err, "HTTP server failed"))
	}
}
