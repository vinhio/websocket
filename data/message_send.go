package data

// MessageSend represents a complete message payload including metadata,
// channel information and the message content for sending through the system
type MessageSend struct {
	// Metadata contains version and server node information
	//Metadata Metadata `json:"metadata"`
	// Channel contains information about the messaging channel
	//Channel Channel `json:"channel"`
	// Message contains the actual message content and metadata
	Message Message `json:"message"`
}
