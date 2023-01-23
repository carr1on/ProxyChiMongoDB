package app

import (
	"context"
	"httpServer/internal/controller"
	"net/http"
	"os"
	"time"

	"log"
	"sync"
)

func App(s *controller.Server, wg *sync.WaitGroup, sig *chan os.Signal) {

	wg.Add(1)
	port := "8081"

	s.MountHandlers()
	server := &http.Server{Addr: ":" + port, Handler: s.Router}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	go func() {
		defer wg.Done()
		<-*sig
		// Shutdown signal with grace period of 10 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 10*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// Run the server
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()

}
