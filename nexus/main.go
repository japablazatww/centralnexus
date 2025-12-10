package main

import (
	"fmt"
	"net/http"

	"github.com/japablazatww/centralnexus/nexus/generated"
)

func main() {
	mux := http.NewServeMux()

	// Register generated handlers
	generated.RegisterHandlers(mux)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := ":8080"
	fmt.Printf("[Nexus] Server listening on %s\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
