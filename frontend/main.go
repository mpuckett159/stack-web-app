// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
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
	router.HandleFunc("/", Ping).Methods("GET")
	router.HandleFunc("/ws", wshandler.GetWS).Methods("GET")
	router.HandleFunc("/ws", wshandler.PostWS).Methods("POST")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	} else {
		fmt.Println("Server successfully started")
	}
}

func Ping(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("🏓 P O N G 🏓"))
	if err != nil {
		return
	}
}
