// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"stack-web-app/frontend/wshandler"
	"stack-web-app/frontend/db"

	"github.com/gorilla/mux"
)

func main() {
	// Set up gorilla mux router handling
	flag.Parse()
	db.Start()
	router := mux.NewRouter()
	s := http.StripPrefix("/public/", http.FileServer(http.Dir("../public/")))
    router.PathPrefix("/public/").Handler(s)
	router.Handle("/", http.FileServer(http.Dir("../public")))
	router.HandleFunc("/favicon.ico", faviconHandler)
	router.HandleFunc("/ws", wshandler.GetWS).Methods("GET")
	router.HandleFunc("/ws", wshandler.PostWS).Methods("POST")

	// Setting some required pieces for DigitalOcean app platform support
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	bindAddr := fmt.Sprintf(":%s", port)
	fmt.Println()
	fmt.Printf("==> Server listening at %s 🚀\n", bindAddr)

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), router)
	if err != nil {
		panic(err)
	}
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../public/favicon.ico")
}