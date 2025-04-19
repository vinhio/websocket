package data

import "time"

// ContentText represents the text content of a message
type ContentText struct {
	// Text is the actual text content/body of the message
	Text string `json:"text"`
}

// ContentFile represents a file attachment in a message
type ContentFile struct {
	// FileType indicates the MIME type or format of the file
	FileType string `json:"file_type"`
	// FileName is the original name of the uploaded file
	FileName string `json:"file_name"`
	// FileSize is the size of the file in bytes
	FileSize int `json:"file_size"`
	// FileHash contains the file's hash/checksum for integrity verification
	FileHash string `json:"file_hash"`
	// FileData contains the base64 encoded file content
	FileData string `json:"file_data"`
	// Caption is an optional text description for the file
	Caption string `json:"caption"`
}

// Reaction tracks emoji responses to messages
type Reaction struct {
	// Emoji is the emoji character/code used for the reaction
	Emoji string `json:"emoji"`
	// Count tracks the total number of users who reacted with this emoji
	Count int `json:"count"`
	// Users contains the IDs of users who added this reaction
	Users []string `json:"users"`
}

// Message represents a chat message with its metadata, content and reactions
type Message struct {
	// ID uniquely identifies the message
	ID string `json:"id"`
	// SenderID identifies the user who sent the message
	SenderID string `json:"sender_id"`
	// Timestamp records when the message was sent
	Timestamp time.Time `json:"timestamp"`
	// Type indicates the kind of message content (e.g. text, file)
	Type string `json:"type"`
	// Content contains the actual message payload as either ContentText or ContentFile
	Content any `json:"content"`
	// Status represents the delivery status of the message (e.g. sent, delivered, read)
	Status string `json:"status"`
	// Reactions tracks emoji responses from users to this message
	Reactions []Reaction `json:"reactions"`
}
