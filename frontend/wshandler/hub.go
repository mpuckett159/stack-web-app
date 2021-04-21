// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wshandler

import (
	"fmt"

	"stack-web-app/frontend/db"
	
	log "github.com/sirupsen/logrus"
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
var HubPool = map[string]*Hub{}

// newHub crates a new hub and registers it with the HubPool global hub table.
func newHub() *Hub {
	// Update context logger
	ContextLogger = ContextLogger.WithFields(log.Fields{
		"module": "hub",
		"function": "newHub",
	})

	// Create new UUID to declare new hub with
	hubId := uuid.New().String()

	// Create new DB table to store users in
	ContextLogger.WithFields(log.Fields{
		"hubId": hubId,
	}).Debug("Creating meeting hub and database table.")
	db.CreateTable(hubId)
	hub := Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		hubId:		hubId,
	}
	ContextLogger.WithFields(log.Fields{
		"hubId": hubId,
		"hub": fmt.Sprintf("%+v", hub),
	}).Debug("Meeting hub and database table successfully created.")

	// Add hub ID to hub pointer map for quick meeting hub lookup
	HubPool[hubId] = &hub
	ContextLogger.WithFields(log.Fields{
		"hub": fmt.Sprintf("%+v", hub),
		"hubPoolMap": HubPool,
	}).Debug("Meeting hub successfully added to HubPool.")

	// Return pointer to the hub object
	return &hub
}

// run is used to start new hubs that have been created.
func (h *Hub) run() {
	// Update context logger
	ContextLogger = ContextLogger.WithFields(log.Fields{
		"module": "hub",
		"function": "run",
	})

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			ContextLogger.WithFields(log.Fields{
				"client": fmt.Sprintf("%+v", client),
				"hub": fmt.Sprintf("%+v", h),
			}).Debug("Client successfully registered to hub.")
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				db.GetOffStack(h.hubId, client.clientId)
				delete(h.clients, client)
				close(client.send)
				ContextLogger.WithFields(log.Fields{
					"client": fmt.Sprintf("%+v", client),
					"hub": fmt.Sprintf("%+v", h),
				}).Debug("Client successfully unregistered from hub, updating stack to rest of group.")
			}
		case message := <-h.broadcast:
			ContextLogger.WithFields(log.Fields{
				"message": fmt.Sprintf("%+v", string(message)),
				"hub": fmt.Sprintf("%+v", h),
			}).Debug("Message being sent to all clients in hub.")
			for client := range h.clients {
				select {
				case client.send <- message:
					ContextLogger.WithFields(log.Fields{
						"client": fmt.Sprintf("%+v", client),
						"hub": fmt.Sprintf("%+v", h),
						"message": fmt.Sprintf("%+v", string(message)),
					}).Debug("Broadcast message being sent to client.")
				default:
					close(client.send)
					delete(h.clients, client)
					ContextLogger.WithFields(log.Fields{
						"client": fmt.Sprintf("%+v", client),
						"hub": fmt.Sprintf("%+v", h),
						"message": fmt.Sprintf("%+v", string(message)),
					}).Debug("Unable to send message to client, successfully unregistered client from hub.")
				}
			}
		}
	}
}