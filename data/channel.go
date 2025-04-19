package data

import "time"

// Participant represents a user participating in a channel with their status information
type Participant struct {
	// UserID uniquely identifies the participant
	UserID string `json:"user_id"`
	// Username is the display name of the participant
	Username string `json:"username"`
	// Status indicates the current online status of the participant (e.g. online, away, offline)
	Status string `json:"status"`
	// JoinedAt records when the participant joined the channel
	JoinedAt time.Time `json:"joined_at"`
}

// Channel represents a messaging channel that participants can interact in
type Channel struct {
	// ID uniquely identifies the channel
	ID string `json:"id"`
	// Type indicates the kind of channel (e.g. direct, group)
	Type string `json:"type"`
	// Participants contains the list of users in this channel
	Participants []Participant `json:"participants"`
	// CreatedAt records when the channel was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt records when the channel was last modified
	UpdatedAt time.Time `json:"updated_at"`
}
