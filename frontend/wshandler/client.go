// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wshandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"stack-web-app/frontend/db"

	log "github.com/sirupsen/logrus"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Unique client ID
	clientId string
}

// The websocket information struct for the a new meeting creation POST method
type WsReturn struct {
	MeetingId	string	`json:"meetingId"`
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	// Update context logger
	ContextLogger = ContextLogger.WithFields(log.Fields{
		"module": "client",
		"function": "readPump",
		"client": fmt.Sprintf("%+v", c),
		"hub": fmt.Sprintf("%+v", c.hub),
	})

	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		// Read next JSON message for user updates
		type userMessage struct {
			TableId	string
			Action	string
			Name	string
		}
		var messageJson userMessage
		err := c.conn.ReadJSON(&messageJson)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				ContextLogger.WithFields(log.Fields{
					"closeError": err.Error(),
				}).Error("Unexpected closure from client.")
			}
			break
		}

		// Put user on/off stack based on action in request
		if(messageJson.Action == "on") {
			db.GetOnStack(messageJson.TableId, c.clientId, messageJson.Name)
		} else if (messageJson.Action == "off") {
			db.GetOffStack(messageJson.TableId, c.clientId)
		}

		// Get current stack back and push to the broadcast message queue
		stackUsers, err := db.ShowCurrentStack(messageJson.TableId)
		if err != nil {
			ContextLogger.WithFields(log.Fields{
				"dbError": err.Error(),
			}).Error("Error getting current meeting stack contents.")
		}
		messageUsers, err := json.Marshal(stackUsers)
		if err != nil {
			ContextLogger.WithFields(log.Fields{
				"dbError": err.Error(),
			}).Error("Error marshalling JSON for response to client.")
		}
		message := bytes.TrimSpace(bytes.Replace(messageUsers, newline, space, -1))
		ContextLogger.WithFields(log.Fields{
			"message": fmt.Sprintf("%+v", string(message)),
		}).Debug("Sending message from client to hub broadcast.")
		c.hub.broadcast <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	// Update context logger
	ContextLogger = ContextLogger.WithFields(log.Fields{
		"module": "client",
		"function": "writePump",
		"client": fmt.Sprintf("%+v", c),
		"hub": fmt.Sprintf("%+v", c.hub),
	})

	// Set ticker
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			ContextLogger.Debug("Sending message to client?")
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				ContextLogger.Debug("Hub has closed this channel, sending update to users.")
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})

				// Sending update to all still connected clients// Get current stack back and push to the broadcast message queue
				db.GetOffStack(c.hub.hubId, c.clientId)
				stackUsers, err := db.ShowCurrentStack(c.hub.hubId)
				if err != nil {
					ContextLogger.WithFields(log.Fields{
						"dbError": err.Error(),
					}).Error("Error getting current meeting stack contents.")
				}
				messageUsers, err := json.Marshal(stackUsers)
				if err != nil {
					ContextLogger.WithFields(log.Fields{
						"dbError": err.Error(),
					}).Error("Error marshalling JSON for response to client.")
				}
				message := bytes.TrimSpace(bytes.Replace(messageUsers, newline, space, -1))
				ContextLogger.WithFields(log.Fields{
					"message": fmt.Sprintf("%+v", string(message)),
				}).Debug("Sending message from client to hub broadcast.")
				c.hub.broadcast <- message
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				ContextLogger.Warning("Error closing writer channel or something?")
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				ContextLogger.Warning("Error pinging the websocket, assuming client is dead and unregistering.")
				c.hub.unregister <- c
				return
			}
		}
	}
}

// GetWS sets up the new WebSocket and connects the client to it. On first connect it also fetches
// the current speaker stack and pushes it out to all the connected clients.
func GetWS(w http.ResponseWriter, r *http.Request) {
	// Update context logger
	ContextLogger = ContextLogger.WithFields(log.Fields{
		"module": "client",
		"function": "GetWS",
	})
	
	// Getting hub ID from http request query params
	hubId := r.URL.Query().Get("meeting_id")
	ContextLogger = ContextLogger.WithField("hubId", hubId)
	var hub *Hub

	// Look for existing meeting hub from ID provided in URL
	if v, ok := HubPool[hubId]; ok {
		hub = v
		ContextLogger = ContextLogger.WithField("hub", fmt.Sprintf("%+v", hub))
	} else {
		ContextLogger.Debug("Meeting not found.")
		return
	}

	// This is to enable local testing for myself. Probably stupid
	_, disableCORS := os.LookupEnv("DISABLEWEBSOCKETORIGINCHECK")
	if disableCORS {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	clientId := uuid.New().String()
	ContextLogger = ContextLogger.WithField("clientId", clientId)
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), clientId: clientId}
	ContextLogger = ContextLogger.WithField("client", fmt.Sprintf("%+v", client))
	client.hub.register <- client
	ContextLogger.Debug("New client successfully registered with hub.")

	// Push current stack out to clients. Should update eventually to only push out to new users somehow
	// Get current stack back and push to the broadcast message queue
	stackUsers, err := db.ShowCurrentStack(hubId)
	if err != nil {
		ContextLogger.WithField("error", err.Error()).Error("Error fetching current speaker stack.")
	}
	messageUsers, err := json.Marshal(stackUsers)
	if err != nil {
		ContextLogger.WithField("error", err.Error()).Error("Error marshaling current stack for client response message.")
	}
	message := bytes.TrimSpace(bytes.Replace(messageUsers, newline, space, -1))
	client.hub.broadcast <- message
	ContextLogger.Debug("Message successfully sent to hub broadcast channel.")

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	ContextLogger.Debug("Starting client read/write goroutines.")
	go client.writePump()
	go client.readPump()
}

// PostWS creates new meeting table in SQLite DB and returns the ID to the client
func PostWS(w http.ResponseWriter, r *http.Request) {
	// Update context logger
	ContextLogger = ContextLogger.WithFields(log.Fields{
		"module": "client",
		"function": "PostWS",
	})

	// Create new hub for meeting and return to be used for client creation
	hub := newHub()
	ContextLogger = ContextLogger.WithField("hub", fmt.Sprintf("%+v", hub))
	ContextLogger.Debug("Starting new hub goroutine.")
	go hub.run()

	// Return new meeting ID to client
	returnBlob := WsReturn{hub.hubId}
	rJson, err := json.Marshal(returnBlob)
	if err != nil {
		ContextLogger.Error("Error marshalling JSON response.")
		return
	}
	ContextLogger.WithField("responseJson", fmt.Sprintf("%+v", returnBlob)).Debug("Sending response to requestor.")
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(rJson)
}