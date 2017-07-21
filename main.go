package main

import (
	"runtime"
	"os"
	"os/signal"
	"log"
	"time"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	// Load config file
	if err := loadConfig("adns.toml"); err != nil {
		log.Fatal(err)
	}

	// Create an sql.DB and check for errors
	db, err := sql.Open("mysql", config.MysqlConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	// Test the connection to the database
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Start dns server
	server := &Server{
		host:     config.DnsBind,
		rTimeout: 5 * time.Second,
		wTimeout: 5 * time.Second,
		db: db,
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