package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"newsletter/internal/infrastructure/workerpool"
	transporthttp "newsletter/transport/http"
)

func main() {
	wp := workerpool.NewWorkerPool(5, 100, &sync.WaitGroup{})
	wp.Start()

	app := transporthttp.NewApp(wp)

	server := &http.Server{
		Addr:    ":8001",
		Handler: app.Routes(),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal()
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Shutting down...")
	server.Shutdown(ctx)

	wp.Shutdown()
	wp.Wait()
}
