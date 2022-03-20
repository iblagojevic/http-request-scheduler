package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/iblagojevic/http-request-scheduler/requestscheduler"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	requestscheduler.SFQ = requestscheduler.InitQueue()
	router := mux.NewRouter()
	router.HandleFunc("/", requestscheduler.HandleIncomingRequests).Methods("POST")

	webserver := &http.Server{
		Addr:    ":9292",
		Handler: router,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGSTOP)

	go func() {
		if err := webserver.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error in webserver: %s\n", err)
		}
	}()
	fmt.Println("Starting to listen on port 9292...")
	// wait for stop signal
	<-stop
	fmt.Println("Webserver stopped.")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// drain queue
		requestscheduler.SFQ.Drain()
		cancel()
	}()

	if err := webserver.Shutdown(ctx); err != nil {
		fmt.Printf("Webserver failed to shutdown gracefully: %+v", err)
	}
}
