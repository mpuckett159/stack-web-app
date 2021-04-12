// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"

	"stack-web-app/frontend/wshandler"
	"stack-web-app/frontend/db"

	"github.com/gorilla/mux"
)

func main() {
	flag.Parse()
	db.Start()
	router := mux.NewRouter()
	s := http.StripPrefix("/public/", http.FileServer(http.Dir("../public/")))
    router.PathPrefix("/public/").Handler(s)
	router.Handle("/", http.FileServer(http.Dir("../public")))
	router.HandleFunc("/ws", wshandler.GetWS).Methods("GET")
	router.HandleFunc("/ws", wshandler.PostWS).Methods("POST")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
