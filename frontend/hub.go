// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"stack-web-app/frontend/db"

	"github.com/google/uuid"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Hub ID so users can join asynchronously
	hubId string
}

// Declare global slice of hub ID to hub pointer map to track existing meeting hubs
var HubPool map[string]*Hub

func newHub() *Hub {
	// Create new UUID to declare new hub with
	hubId := uuid.New().String()

	// Create new DB table to store users in
	db.CreateTable(hubId)
	hub := Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		hubId:		hubId,
	}

	// Add hub ID to hub pointer map for quick meeting hub lookup
	HubPool[hubId] = &hub

	// Return pointer to the hub object
	return &hub
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}