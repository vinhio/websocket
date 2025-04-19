package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/gflydev/core/log"
	"time"
	"ws/data"
	"ws/websocket"
)

const (
	// Time allowed writing a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed reading the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10240
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

		// Process the message based on its format and content
		msgStr := string(message)

		// Legacy support for the "/switch" command
		if len(msgStr) > 8 && msgStr[:8] == "/switch " {
			channelID := msgStr[8:]
			log.Infof("Client %s requesting channel switch to %s (legacy format)", c.conn.RemoteAddr(), channelID)

			// Get the hub for the specified channel or create a new one if it doesn't exist
			hub := manager.GetHub(channelID)
			if hub == nil {
				hub = manager.CreateChannelHub(channelID, channelID)
			}

			// Switch the client to the new channel without closing the WebSocket connection
			c.SwitchChannel(hub)
		} else {
			// Try to parse the message as an ActionMessage first
			var actionMsg data.ActionMessage
			err = json.Unmarshal(message, &actionMsg)

			if err == nil && actionMsg.Action.Type != "" {
				// Process the message based on its action type
				c.handleAction(actionMsg)
			} else {
				// Try to parse as legacy MessageSend format
				var msgSend data.MessageSend
				err = json.Unmarshal(message, &msgSend)

				if err != nil {
					// If not valid JSON, create a new message with the text content
					msgSend = data.MessageSend{
						Message: data.Message{
							ID:        generateID(),
							SenderID:  c.id,
							Timestamp: time.Now(),
							Type:      "text",
							Content: data.ContentText{
								Text: msgStr,
							},
							Status:    "sent",
							Reactions: []data.Reaction{},
						},
					}
				}

				// Add channel information if available
				if c.hub.name != "" {
					// Add channel name to the message content for display purposes
					if textContent, ok := msgSend.Message.Content.(data.ContentText); ok {
						textContent.Text = "[" + c.hub.name + "] " + textContent.Text
						msgSend.Message.Content = textContent
					}
				}

				// Convert the message to JSON
				jsonMessage, err := json.Marshal(msgSend)
				if err != nil {
					log.Errorf("Error marshaling message to JSON: %v", err)
					return
				}

				c.hub.broadcast <- jsonMessage
			}
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
			// NOTE: Will get the error inside c.conn.RemoteAddr() for everytime the client disconnects
			// error detail `panic: runtime error: invalid memory address or nil pointer dereference`
			//log.Debug("Ping", c.conn.RemoteAddr())

			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) pongHandler(string) error {
	// NOTE: Will get the error inside c.conn.RemoteAddr() for everytime the client disconnects
	// error detail `panic: runtime error: invalid memory address or nil pointer dereference`
	//log.Debug("Pong", c.conn.RemoteAddr())

	// Update the read deadline when a Pong message is received.
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))

	if err != nil {
		log.Error("Error setting read deadline: ", err)
	}

	return err
}

// generateID creates a random ID string for messages
func generateID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		// If random generation fails, use timestamp as fallback
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(bytes)
}

// handleAction processes an ActionMessage based on its action type.
// It handles various action types like switching channels, listing channels, sending messages, etc.
//
// Parameters:
// - actionMsg (data.ActionMessage): The action message to process.
func (c *Client) handleAction(actionMsg data.ActionMessage) {
	switch actionMsg.Action.Type {
	case data.ActionSwitchChannel:
		// Handle channel switching
		if switchData, ok := actionMsg.Action.Data.(map[string]interface{}); ok {
			if channelID, ok := switchData["channel_id"].(string); ok && channelID != "" {
				log.Infof("Client %s requesting channel switch to %s", c.conn.RemoteAddr(), channelID)

				// Get the hub for the specified channel or create a new one if it doesn't exist
				hub := manager.GetHub(channelID)
				if hub == nil {
					hub = manager.CreateChannelHub(channelID, channelID)
				}

				// Switch the client to the new channel
				c.SwitchChannel(hub)
				return
			}
		}
		log.Errorf("Invalid channel switch data: %v", actionMsg.Action.Data)

	case data.ActionListChannels:
		// Handle channel listing
		channels := make([]data.Channel, 0)

		// Collect all available channels
		for id, _ := range manager.poolHub {
			channels = append(channels, data.Channel{
				ID:           id,
				Type:         "group",
				Participants: []data.Participant{},
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			})
		}

		// Create a response message
		response := data.ActionMessage{
			Metadata: data.Metadata{
				Version:   "1.0",
				Timestamp: time.Now(),
			},
			Action: data.Action{
				Type: data.ActionListChannels,
				Data: data.ChannelListData{
					Channels: channels,
				},
			},
		}

		// Send the response only to the requesting client
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			log.Errorf("Error marshaling channel list to JSON: %v", err)
			return
		}

		c.send <- jsonResponse

	case data.ActionSendMessage:
		// Handle message sending
		if msgData, ok := actionMsg.Action.Data.(map[string]interface{}); ok {
			// Extract the message from the action data
			msgBytes, err := json.Marshal(msgData)
			if err != nil {
				log.Errorf("Error marshaling message data: %v", err)
				return
			}

			var messageData data.MessageSendData
			err = json.Unmarshal(msgBytes, &messageData)
			if err != nil {
				log.Errorf("Error unmarshaling message data: %v", err)
				return
			}

			// Create a MessageSend structure
			msgSend := data.MessageSend{
				Metadata: actionMsg.Metadata,
				Channel:  actionMsg.Channel,
				Message:  messageData.Message,
			}

			// Add channel information if available
			if c.hub.name != "" && msgSend.Message.Type == "text" {
				if textContent, ok := msgSend.Message.Content.(map[string]interface{}); ok {
					if text, ok := textContent["text"].(string); ok {
						textContent["text"] = "[" + c.hub.name + "] " + text
					}
				}
			}

			// Convert the message to JSON
			jsonMessage, err := json.Marshal(msgSend)
			if err != nil {
				log.Errorf("Error marshaling message to JSON: %v", err)
				return
			}

			// Broadcast the message to all clients in the hub
			c.hub.broadcast <- jsonMessage
		}

	case data.ActionCreateChannel:
		// Handle channel creation
		if createData, ok := actionMsg.Action.Data.(map[string]interface{}); ok {
			if channelData, ok := createData["channel"].(map[string]interface{}); ok {
				if channelID, ok := channelData["id"].(string); ok && channelID != "" {
					// Create a new hub for the channel
					_ = manager.CreateChannelHub(channelID, channelID)

					// Send a confirmation message
					response := data.ActionMessage{
						Metadata: data.Metadata{
							Version:   "1.0",
							Timestamp: time.Now(),
						},
						Action: data.Action{
							Type: data.ActionCreateChannel,
							Data: data.ChannelCreateData{
								Channel: data.Channel{
									ID:           channelID,
									Type:         "group",
									Participants: []data.Participant{},
									CreatedAt:    time.Now(),
									UpdatedAt:    time.Now(),
								},
							},
						},
					}

					jsonResponse, err := json.Marshal(response)
					if err != nil {
						log.Errorf("Error marshaling channel creation response to JSON: %v", err)
						return
					}

					c.send <- jsonResponse
					return
				}
			}
		}
		log.Errorf("Invalid channel creation data: %v", actionMsg.Action.Data)

	default:
		// For unhandled action types, just log a message
		log.Infof("Unhandled action type: %s", actionMsg.Action.Type)
	}
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

	// Send a JSON notification to the client about the channel switch
	switchMsg := data.MessageSend{
		Message: data.Message{
			ID:        generateID(),
			SenderID:  "system",
			Timestamp: time.Now(),
			Type:      "text",
			Content: data.ContentText{
				Text: "Switched to channel: " + c.hub.name,
			},
			Status:    "sent",
			Reactions: []data.Reaction{},
		},
	}

	jsonMessage, err := json.Marshal(switchMsg)
	if err != nil {
		log.Errorf("Error marshaling channel switch notification to JSON: %v", err)
		return
	}

	c.send <- jsonMessage
}
