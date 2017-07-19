package main

import (
	"runtime"
	"os"
	"os/signal"
	"log"
	"time"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	if err := loadConfig("adns.toml"); err != nil {
		log.Fatal(err)
	}

	// todo start logger

	// todo check mysql server

	server := &Server{
		host:     config.DnsBind,
		rTimeout: 5 * time.Second,
		wTimeout: 5 * time.Second,
	}

	server.Run()

	// todo start api server

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	forever:
	for {
		select {
		case <-sig:
			log.Printf("signal received, stopping\n")
			break forever
		}
	}
}