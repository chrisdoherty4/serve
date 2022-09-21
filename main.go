package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Configure flags using the flag singleton.
	dir := flag.String("d", ".", "The directory to serve files from.")
	address := flag.String("a", ":8080", "The address to listen on.")
	silent := flag.Bool("s", false, "Silence server logging.")
	flag.Parse()

	if *silent {
		log.SetOutput(io.Discard)
	}

	log.Printf("Serving files from %v", *dir)

	mux := http.NewServeMux()

	// Add a handler that logs requests delegating file serving to the standard libraries FileServer.
	fileServer := http.FileServer(http.Dir(*dir))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, Status: 200}
		fileServer.ServeHTTP(recorder, r)
		log.Printf("%v %v %v", r.Method, recorder.Status, r.URL.Path)
	}))

	// Create a server and launch in a go routine so we can gracefully shutdown when signaled.
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

// statusRecorder records the status of a request written with WriteHeader.
type statusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}
