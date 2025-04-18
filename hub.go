package main

func newHub(name string) *Hub {
	// Creates and returns a new Hub instance.
	//
	// Parameters:
	// - name (string): The name of the channel/room.
	//
	// Logic:
	// 1. Initializes and returns a pointer to a new Hub instance.
	// 2. Sets up the following fields for the Hub instance:
	//	- broadcast: A channel for receiving inbound messages from clients to be broadcast to other clients.
	//	- register: A channel for handling client registration requests.
	//	- unregister: A channel for handling client unregistration requests.
	//	- clients: A map to manage and store the active clients.
	//	- name: The name of the channel/room.
	//
	// Returns:
	// - *Hub: A pointer to a newly created Hub instance.
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		name:       name,
	}
}

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

	// Name of the channel/room
	name string
}

// IsEmpty Check if the hub's client map is empty.
//
// Parameters:
// - None.
//
// Logic:
// 1. Returns true if the clients map is empty, otherwise false.
func (h *Hub) IsEmpty() bool {
	return len(h.clients) == 0
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
			// - If it does, remove it from the map.
			// - IMPORTANT: We don't close the send channel here anymore to support channel switching.
			//   This is a key change that allows clients to switch channels without disconnecting.
			//   Previously, the send channel was closed here, which would break the WebSocket connection
			//   when a client switched channels.
			// - The send channel will be closed in the client's readPump when the connection is actually closed
			//   (i.e., when the client disconnects from the server, not just when switching channels).
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				// Don't close the send channel here to support channel switching
				// close(client.send) - This would break channel switching
			}
		case message := <-h.broadcast: // Parameter: message ([]byte) - A message received from a client to be broadcast to all clients.
			// Logic:
			// - Loop through all currently registered clients.
			// - Attempt to send the message through each client's send channel.
			// - If a client's send channel is full (default case), close the client's channel and unregister the client by removing it from the hub.
			for client := range h.clients {
				select {
				case client.send <- message: // Successfully send the message.
				default: // Failed to send a message (channel full or disconnected).
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
