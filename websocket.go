package main

import (
	"github.com/gflydev/core"
	"github.com/gflydev/core/log"
	"github.com/gflydev/core/try"
	"github.com/gflydev/core/utils"
	"github.com/valyala/fasthttp"
	"net/url"
	"strings"
	"ws/websocket"
)

// ====================================================================
// ============================ Structure =============================
// ====================================================================

const (
	// DefaultHubID is the identifier for the default Hub instance.
	DefaultHubID = "default"
)

// PoolHub is a map that associates a unique string identifier with a pointer to a Hub instance.
// It is used to manage multiple Hub instances within the application.
type PoolHub map[string]*Hub

// Manager is responsible for managing multiple Hub instances.
// It contains a poolHub, which is a map of Hub instances identified by unique string keys.
type Manager struct {
	poolHub PoolHub // A map to store and manage Hub instances.
}

// NewManager creates and initializes a new Manager instance.
//
// Returns:
// - *Manager: A pointer to the newly created Manager instance with an initialized poolHub.
func NewManager() *Manager {
	return &Manager{
		poolHub: make(PoolHub),
	}
}

// createDefaultHub creates and initializes the default hub for the manager.
//
// This method performs the following steps:
// 1. Creates a new Hub instance using the `newHub` function.
// 2. Adds the newly created Hub to the `poolHub` map with the key `DefaultHubID`.
// 3. Starts the Hub's `run` method in a separate goroutine to handle client connections and messages.
//
// Returns:
// - *Hub: A pointer to the newly created default Hub instance.
func (m *Manager) createDefaultHub() *Hub {
	hub := newHub()
	manager.poolHub[DefaultHubID] = hub
	go hub.run()

	return hub
}

// GetHub retrieves a Hub instance by its ID.
//
// Parameters:
// - id (string): The unique identifier of the Hub to retrieve.
//
// Returns:
// - *Hub: A pointer to the Hub instance if found, or nil if no Hub exists with the given ID.
func (m *Manager) GetHub(id string) *Hub {
	if hub, ok := m.poolHub[id]; ok {
		return hub
	}
	return nil
}

// SetHub adds a new Hub instance to the poolHub map if it does not already exist.
//
// Parameters:
// - id (string): The unique identifier for the Hub.
// - hub (*Hub): A pointer to the Hub instance to be added.
//
// Logic:
// 1. Checks if the given id already exists in the poolHub map.
// 2. If it does not exist, adds the Hub instance to the map.
func (m *Manager) SetHub(id string, hub *Hub) {
	if _, ok := m.poolHub[id]; !ok {
		m.poolHub[id] = hub
	}
}

// DeleteHub removes a Hub instance from the poolHub map if it is empty.
//
// Parameters:
// - id (string): The unique identifier for the Hub to be deleted.
//
// Logic:
// 1. Checks if the given id exists in the poolHub map.
// 2. If the Hub exists and is empty (no active clients), deletes it from the map.
// 3. If the Hub is not empty, logs a warning message and does not delete it.
func (m *Manager) DeleteHub(id string) {
	if hub, ok := m.poolHub[id]; ok {
		if hub.IsEmpty() {
			delete(m.poolHub, id)
		} else {
			log.Warn("Hub is not empty, cannot delete")
		}
	}
}

// ====================================================================
// =========================== gFly Handler ===========================
// ====================================================================

// NewWSHandler As a constructor to create a WebSocket handler.
func NewWSHandler() *WSHandler {
	return &WSHandler{}
}

type WSHandler struct {
	core.Page
}

func (m *WSHandler) Handle(c *core.Ctx) error {
	ServeWS(c.Root())

	return nil
}

// ====================================================================
// ====================== Websocket integration =======================
// ====================================================================

var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  10240,
	WriteBufferSize: 10240,
	// Apply the Origin Checker
	CheckOrigin: checkOrigin,
}

// checkOrigin will check origin and return true if its allowed
func checkOrigin(r *fasthttp.RequestCtx) bool {
	// Grab the request origin
	originUrl := string(r.Request.Header.Peek("Origin"))
	appUrl := utils.Getenv[string]("APP_URL", "")

	originInstance, err := url.Parse(originUrl)
	if err != nil {
		return false
	}
	appInstance, err := url.Parse(appUrl)
	if err != nil {
		return false
	}

	originUrl = strings.TrimPrefix(originInstance.Hostname(), "www.")
	appUrl = strings.TrimPrefix(appInstance.Hostname(), "www.")

	log.Debug(originUrl, appUrl)

	switch originUrl {
	case appUrl:
		return true
	default:
		return false
	}
}

// manager is the global Manager instance that manages all Hub instances.
var manager = NewManager()

func init() {
	// Initialize the default hub
	manager.createDefaultHub()
}

// ServeWS handles websocket requests from the peer.
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
func ServeWS(ctx *fasthttp.RequestCtx) {
	// Get the hub ID from the request URL
	hub := manager.GetHub(DefaultHubID)

	try.Perform(func() {
		err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
			client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
			client.hub.register <- client

			go client.writePump()
			client.readPump()
		})

		if err != nil {
			log.Error(err)
		}
	}).Catch(func(e try.E) {
		log.Error("Error in goroutine: ", e)
	})
}
