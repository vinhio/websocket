package data

import "time"

// ServerNode represents a server node in the cluster with its identification and status information
type ServerNode struct {
	// ID uniquely identifies the server node
	ID string `json:"id"`
	// Region indicates the geographic location of the server
	Region string `json:"region"`
	// Load represents the current CPU load average of the server node (0.0 - 1.0)
	Load float64 `json:"load"`
}

// Metadata contains version and timing information about messages
type Metadata struct {
	// Version indicates the message protocol version
	Version string `json:"version"`
	// Timestamp records when the message was created
	Timestamp time.Time `json:"timestamp"`
	// ServerNode contains information about the server handling the message
	ServerNode ServerNode `json:"server_node"`
}
