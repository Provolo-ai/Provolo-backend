package routes

import (
	"fmt"
	"log"
	"net/http"
	"provolo-api/internal/types"
	"time"
)

// StartServer starts the HTTP server with the given configuration
func StartServer(config *types.Config) error {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      SetupRoutes(config),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("Starting server on port %d", config.Port)

	return server.ListenAndServe()
}
