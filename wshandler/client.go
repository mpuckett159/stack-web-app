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

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
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
	MeetingId string `json:"meetingId"`
}

// Client message formats
//
// MeetingId - links clients to meeting hub struct
// Action - the action that will be performed. This will be parsed out on the client side
// ClientId - the source client ID for the action, which will inform some actions on the client side app
type UserMessage struct {
	MeetingId string `json:"meetingId"`
	Action    string `json:"action"`
	ClientId  string `json:"clientId"`
}

// Meeting creation body format
type MeetingCreationMessage struct {
	ModActions []string `json:"actions"`
}

// Abstrcts marshaling and sending a JSON message to a meeting hub
func (c *Client) broadcastMessage(messageJson UserMessage) {
	messageUsers, err := json.Marshal(messageJson)
	if err != nil {
		ContextLogger.WithFields(log.Fields{
			"dbError": err.Error(),
		}).Error("Error marshalling JSON for response to client.")
	}

	// Verify that mod actions are coming from the actual mod
	message := bytes.TrimSpace(bytes.Replace(messageUsers, newline, space, -1))
	ContextLogger.WithFields(log.Fields{
		"message": fmt.Sprintf("%+v", string(message)),
	}).Debug("Sending message from client to hub broadcast.")
	c.hub.broadcast <- message
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	// Update context logger
	ContextLogger = ContextLogger.WithFields(log.Fields{
		"module":   "client",
		"function": "readPump",
		"client":   fmt.Sprintf("%+v", c),
		"hub":      fmt.Sprintf("%+v", c.hub),
	})

	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		ContextLogger.Error("Error setting client connection read deadline")
	}
	c.conn.SetPongHandler(func(string) error {
		err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			ContextLogger.Error("Error setting client connection read deadline")
		}
		return nil
	})
	for {
		// Read next JSON message for user updates
		var messageJson UserMessage
		err := c.conn.ReadJSON(&messageJson)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				ContextLogger.WithFields(log.Fields{
					"closeError": err.Error(),
				}).Error("Unexpected closure from client.")
			}
			break
		}

		// Passing userMessage on to rest of clients for client side handling
		// Checks for client actions vs mod action map and logs error if non mod user
		// tries to perform a mod action, but ignores the attempt.
		if _, ok := c.hub.modActions[messageJson.Action]; (ok && c == c.hub.mod) {
			c.broadcastMessage(messageJson)
		} else if _, ok := c.hub.modActions[messageJson.Action]; (!ok){
			c.broadcastMessage(messageJson)
		} else {
			ContextLogger.Error("A client who was not a mod tried to send mod action message.")
		}
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
		"module":   "client",
		"function": "writePump",
		"client":   fmt.Sprintf("%+v", c),
		"hub":      fmt.Sprintf("%+v", c.hub),
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
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				ContextLogger.Error("Error setting write deadline for client.")
			}
			if !ok {
				// The hub closed the channel.
				ContextLogger.Debug("Hub has closed this channel, sending update to users.")

				// We don't care if the write message fails, so to appease the golangci-lint gods we just log err out to nothing
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})

				// Sends a message informing the other clients that this client is leaving and can be removed from their data.
				meetingUserLeave := UserMessage{
					c.hub.hubId,
					"leave",
					c.clientId,
				}
				c.broadcastMessage(meetingUserLeave)
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, err = w.Write(message)
			if err != nil {
				ContextLogger.Error("Error writing back message to rest of clients after client connection closed.")
			}

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, err = w.Write(newline)
				if err != nil {
					ContextLogger.Error("Error writing back newline back to rest of clients after client connection closed.")
				}
				_, err = w.Write(<-c.send)
				if err != nil {
					ContextLogger.Error("Error sending message to rest of clients after client connection closed.")
				}
			}

			if err := w.Close(); err != nil {
				ContextLogger.Warning("Error closing writer channel or something?")
				return
			}
		case <-ticker.C:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				ContextLogger.Error("Error setting write deadline for client connection.")
			}
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
		"module":   "client",
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

	// Upgrade the client connection to a WebSocket and register the new client in the hub
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

	// Send new user update to clients
	newUserMessage := UserMessage{
		hub.hubId,
		"newuser",
		client.clientId,
	}
	messageUsers, err := json.Marshal(newUserMessage)
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
		"module":   "client",
		"function": "PostWS",
	})

	// Parse the incoming message data to create a new meeting room
	var inMessage MeetingCreationMessage
	err := json.NewDecoder(r.Body).Decode(&inMessage)
	if err != nil {
		ContextLogger.Debug("There was an error parsing the incoming message data.")
	}

	// Create new hub for meeting and return to be used for client creation
	hub := newHub(inMessage.ModActions)
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
	_, err = w.Write(rJson)
	if err != nil {
		ContextLogger.Error("Error writing response back to web session after user requested new meeting.")
	}
}
