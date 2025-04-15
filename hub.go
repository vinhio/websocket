// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

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
}

func newHub() *Hub {
	// Creates and returns a new Hub instance.
	//
	// Parameters:
	// - None.
	//
	// Logic:
	// 1. Initializes and returns a pointer to a new Hub instance.
	// 2. Sets up the following fields for the Hub instance:
	//	- broadcast: A channel for receiving inbound messages from clients to be broadcast to other clients.
	//	- register: A channel for handling client registration requests.
	//	- unregister: A channel for handling client unregistration requests.
	//	- clients: A map to manage and store the active clients.
	//
	// Returns:
	// - *Hub: A pointer to a newly created Hub instance.
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	// Logic:
	// 1. Continuously listen for incoming events on one of the hub's channels using a select statement.
	// 2. Handle the specific event type and update hub's state.

	for {
		select {
		case client := <-h.register: // Parameter: client (*Client) - A new client attempting to connect to the hub.
			// Logic:
			// - Mark the client as registered by adding it to the hub's client map.
			h.clients[client] = true
		case client := <-h.unregister: // Parameter: client (*Client) - A client attempting to disconnect from the hub.
			// Logic:
			// - Check if the client exists in the client map.
			// - If it does, remove it from the map and close its send channel to stop further communication.
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast: // Parameter: message ([]byte) - A message received from a client to be broadcast to all clients.
			// Logic:
			// - Loop through all currently registered clients.
			// - Attempt to send the message through each client's send channel.
			// - If a client's send channel is full (default case), close the client's channel and unregister the client by removing it from the hub.
			for client := range h.clients {
				select {
				case client.send <- message: // Successfully send the message.
				default: // Failed to send message (channel full or disconnected).
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
