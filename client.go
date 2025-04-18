package main

import (
	"bytes"
	"github.com/gflydev/core/log"
	"time"

	"ws/websocket"
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

// Client is an intermediary between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Client ID for tracking across channel switches
	id string
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
//
// A goroutine running readPump is started for each client connection.
//
// Parameters:
// - None (method receiver `c *Client` is the client instance running this function).
//
// Logic:
// 1. A deferred function is executed at the end of the method:
//   - Sends the client instance to the hub's unregister channel to remove it from the active client list.
//   - Closes the websocket connection.
//
// 2. Sets the read limit for the websocket connection using `maxMessageSize`.
//   - Ensures that incoming messages do not exceed this size.
//
// 3. Updates the read deadline based on the `pongWait` duration to detect timeouts on the websocket.
//
// 4. Configures a pong handler function for the websocket connection:
//   - Extends the read deadline each time a Pong message is received from the peer.
//   - Uses `pongWait` to maintain an active connection.
//
// 5. Enters an infinite loop to continuously read messages from the websocket:
//
//   - Reads a message from the websocket connection.
//
//   - If an error occurs while reading:
//
//   - Checks if the error is an unexpected close error (e.g., peer disconnection).
//
//   - Logs the error and exits the loop.
//
//   - If no error occurs:
//
//   - Trims spaces and newlines from the message to sanitize it.
//
//   - Sends the sanitized message to the hub's broadcast channel for distribution to other clients.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		err := c.conn.Close()
		if err != nil {
			return
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)

	// Set pong handler to update read deadline on pong message.
	if err := c.pongHandler(""); err != nil {
		return
	}
	c.conn.SetPongHandler(c.pongHandler)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("Error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// Check if this is a channel switch command
		// Format: "/switch channelID"
		// This is a special command that allows clients to switch channels without disconnecting
		// their WebSocket connection. When a client sends this command, the server will move
		// the client from its current hub to the specified hub, maintaining the WebSocket connection.
		msgStr := string(message)
		if len(msgStr) > 8 && msgStr[:8] == "/switch " {
			channelID := msgStr[8:]
			log.Infof("Client %s requesting channel switch to %s", c.conn.RemoteAddr(), channelID)

			// Get the hub for the specified channel or create a new one if it doesn't exist
			hub := manager.GetHub(channelID)
			if hub == nil {
				hub = manager.CreateChannelHub(channelID, channelID)
			}

			// Switch the client to the new channel without closing the WebSocket connection
			// This is the key part of the implementation that allows clients to switch channels
			// without disconnecting and reconnecting
			c.SwitchChannel(hub)
		} else {
			// Regular message, broadcast to the current hub
			// Prepend channel name to message if available
			if c.hub.name != "" {
				channelPrefix := []byte("[" + c.hub.name + "] ")
				message = append(channelPrefix, message...)
			}
			c.hub.broadcast <- message
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
// It ensures there is only one writer routine per connection and handles ping messages to keep the connection alive.
//
// Parameters:
// - None (method receiver `c *Client` is the client instance running this function).
//
// Logic:
// 1. A `ticker` is initialized to send periodic pings to the websocket connection.
//   - The ticker interval is based on the `pingPeriod`.
//
// 2. A deferred function is set up to:
//   - Stop the ticker when the function exits.
//   - Close the websocket connection to release resources.
//
// 3. An infinite loop processes messages using a `select` statement:
//
//   - Case 1: A message is received from the `send` channel of the client.
//
//   - The write deadline for the websocket is updated based on `writeWait`.
//
//   - If the `send` channel is closed (`ok == false`), the connection is terminated
//     using a websocket `CloseMessage`, and the loop exits.
//
//   - Otherwise:
//
//   - A new `TextMessage` writer is created for the websocket connection.
//
//   - The received message is written to the writer.
//
//   - Any queued messages in the `send` channel are added to the same writer in order and separated by newlines.
//
//   - The writer is closed to finalize the message. Failure to close the writer will terminate the loop.
//
//   - Case 2: The ticker signals a timer event.
//
//   - The write deadline for the websocket is updated based on `writeWait`.
//
//   - A `PingMessage` is sent on the websocket to keep the connection alive.
//
//   - If this fails, the loop exits, and the connection is terminated.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		err := c.conn.Close()
		if err != nil {
			return
		}
	}()
	for {
		select {
		case message, ok := <-c.send:
			var err error

			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write(newline)
				_, _ = w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			log.Debug("Ping", c.conn.RemoteAddr())
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) pongHandler(string) error {
	log.Debug("Pong", c.conn.RemoteAddr())

	// Update the read deadline when a Pong message is received.
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))

	if err != nil {
		log.Error("Error setting read deadline: ", err)
	}

	return err
}

// SwitchChannel switches the client to a new hub without closing the WebSocket connection.
// It unregisters the client from its current hub and registers it with the new hub.
//
// This method is a key part of the implementation that allows clients to switch channels
// without disconnecting and reconnecting their WebSocket connection. It maintains the
// existing connection while changing the hub that the client is associated with.
//
// Parameters:
// - newHub (*Hub): The new hub to register the client with.
//
// Logic:
// 1. Unregister the client from its current hub.
// 2. Update the client's hub reference to the new hub.
// 3. Register the client with the new hub.
func (c *Client) SwitchChannel(newHub *Hub) {
	if c.hub.name == newHub.name {
		return // Already in this hub
	}

	// Unregister from the current hub
	c.hub.unregister <- c

	// Update hub reference
	c.hub = newHub

	// Register with a new hub
	c.hub.register <- c

	// Send a notification to the client about the channel switch
	message := []byte("Switched to channel: " + c.hub.name)
	c.send <- message
}
