package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Configure flags using the flag singleton.
	dir := flag.String("dir", ".", "The directory to serve files from. Defaults to the current dir.")
	address := flag.String("address", ":8080", "The address to listen on. Defaults to :8080.")
	flag.Parse()

	log.Printf("Serving files from %v", *dir)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(*dir)))

	server := http.Server{
		Addr:    *address,
		Handler: mux,
	}

	log.Printf("Starting server on %v", *address)
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Received unexpected error: %v", err)
		}
	}()

	// Listen for signals so we can gracefully shutdown.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Print("Shutting down server")

	// Create a context to ensure graceful shutdown does in-fact terminate and then shut th server
	// down.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Received unexpected error during shutdown: %v", err)
	}
}
