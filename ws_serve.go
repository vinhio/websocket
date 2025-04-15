package main

import (
	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
	"log"
)

var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Apply the Origin Checker
	CheckOrigin: checkOrigin,
}

// checkOrigin will check origin and return true if its allowed
func checkOrigin(r *fasthttp.RequestCtx) bool {

	// Grab the request origin
	origin := string(r.Request.Header.Peek("Origin"))

	switch origin {
	// Update this to HTTPS | HTTP
	case "http://localhost:8080",
		"https://localhost:8080":
		return true
	default:
		return false
	}
}

// serveWs handles websocket requests from the peer.
//
// Parameters:
// - ctx: The `fasthttp.RequestCtx` representing the current HTTP request context.
//   - Used for parsing the request and initializing the websocket connection.
//
// - hub: Pointer to the `Hub` instance that manages client connections and message broadcasting.
//
// Logic:
// 1. Attempts to upgrade an incoming HTTP request to a websocket connection using the `upgrader.Upgrade` method.
//   - If the upgrade fails, logs the error and exits the function.
//
// 2. On successful connection upgrade:
//
//   - A new `Client` instance is created:
//
//   - `hub` is set to the `Hub` instance for managing this client.
//
//   - `conn` is set to the newly established websocket connection.
//
//   - `send` is initialized as a buffered channel for sending messages to the client.
//
//   - The new `Client` is registered with the `Hub` by sending it to the `register` channel.
//
//   - Two goroutines are started to handle the client's websocket connection:
//
//   - `writePump`: Responsible for sending messages to the client and handling pings.
//
//   - `readPump`: Responsible for reading messages from the client.
//
// 3. If an error occurs during the websocket upgrade, it is logged using the `log.Println` function.
func serveWs(ctx *fasthttp.RequestCtx, hub *Hub) {
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
		client.hub.register <- client

		go client.writePump()
		client.readPump()
	})

	if err != nil {
		log.Println(err)
	}
}
