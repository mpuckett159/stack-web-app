// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"stack-web-app/db"
	"stack-web-app/wshandler"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Get any necessary environment variables
	_, debug := os.LookupEnv("DEBUG")

	// Set logger settings
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	// Set up gorilla mux router handling
	flag.Parse()
	db.Start()
	go wshandler.PruneEmptyMeetings()
	router := mux.NewRouter()
	router.HandleFunc("/", wshandler.GetWS).Methods("GET")
	router.HandleFunc("/", wshandler.PostWS).Methods("POST")

	// Set up request logging
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)

	// Setting some required pieces for DigitalOcean app platform support
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	log.WithFields(log.Fields{
		"port": port,
	}).Info(fmt.Sprintf("==> Server listening on port %s 🚀", port))

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), loggedRouter)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Fatal error with HTTP server")
	}
}
