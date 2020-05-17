package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hariadivicky/nano"
)

func main() {
	app := nano.New()

	// simple endpoint to print hello world.
	app.GET("/", func(c *nano.Context) {
		c.String(http.StatusOK, "hello world\n")
	})

	// below is logic to gracefuly shutdown the web server.
	// done channel is used to notify when the shutting down process is complete.
	done := make(chan struct{})
	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// create server from http std package
	server := &http.Server{
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  30 * time.Second,
		Handler:      app, // append nano app as server handler.
		Addr:         ":8000",
	}

	go shutdownHandler(server, shutdown, done)

	log.Println("server running")
	server.ListenAndServe()

	// waiting web server to complete shutdown.
	<-done
	log.Println("server closed")
}

// shutdownHandler do the graceful shutdown to web server.
// when shutdown signal occured, it will wait all active request to completly receive their responses.
// we will wait all unfinished request until 30 seconds.
func shutdownHandler(server *http.Server, shutdown <-chan os.Signal, done chan struct{}) {
	// waiting for shutdown signal.
	<-shutdown
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("could not shutdown server: %v", err)
	}

	close(done)
}
