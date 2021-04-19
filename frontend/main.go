// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"stack-web-app/frontend/wshandler"
	"stack-web-app/frontend/db"

    "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	// Set up gorilla mux router handling
	flag.Parse()
	db.Start()
	router := mux.NewRouter()
	router.HandleFunc("/ws", wshandler.GetWS).Methods("GET")
	router.HandleFunc("/ws", wshandler.PostWS).Methods("POST")

	// Set up request logging
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)

	// Setting some required pieces for DigitalOcean app platform support
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	bindAddr := fmt.Sprintf(":%s", port)
	fmt.Println()
	fmt.Printf("==> Server listening at %s 🚀\n", bindAddr)

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), loggedRouter)
	if err != nil {
		panic(err)
	}
}