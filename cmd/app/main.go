package main

import (
	"httpServer/internal/controller"
	"httpServer/internal/controller/app"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)


func main() {

	var wg = sync.WaitGroup{}
	s := controller.CreateNewServer()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	app.App(s, &wg, &sig)


	wg.Wait()
	log.Print("we have a way out")
	log.Print("we leave")

}