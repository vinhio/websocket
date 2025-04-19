package data

// ActionType defines the type of action being performed in the WebSocket communication
type ActionType string

// Define constants for different action types
const (
	// Channel-related actions
	ActionSwitchChannel ActionType = "switch_channel"
	ActionListChannels  ActionType = "list_channels"
	ActionCreateChannel ActionType = "create_channel"
	ActionJoinChannel   ActionType = "join_channel"
	ActionLeaveChannel  ActionType = "leave_channel"

	// Message-related actions
	ActionSendMessage   ActionType = "send_message"
	ActionEditMessage   ActionType = "edit_message"
	ActionDeleteMessage ActionType = "delete_message"
	ActionReactMessage  ActionType = "react_message"

	// User-related actions
	ActionUserJoin     ActionType = "user_join"
	ActionUserLeave    ActionType = "user_leave"
	ActionUserTyping   ActionType = "user_typing"
	ActionUserPresence ActionType = "user_presence"
)

// Action represents the action being performed in a WebSocket message
type Action struct {
	// Type indicates the kind of action being performed
	Type ActionType `json:"type"`

	// Data contains the payload for the action
	// The structure of Data depends on the Type
	Data interface{} `json:"data"`
}

// ChannelSwitchData contains data for switching channels
type ChannelSwitchData struct {
	// ChannelID is the ID of the channel to switch to
	ChannelID string `json:"channel_id"`
}

// ChannelListData contains data for listing channels
type ChannelListData struct {
	// Channels is the list of available channels
	Channels []Channel `json:"channels"`
}

// ChannelCreateData contains data for creating a new channel
type ChannelCreateData struct {
	// Channel contains the details of the channel to create
	Channel Channel `json:"channel"`
}

// MessageSendData contains data for sending a message
type MessageSendData struct {
	// Message contains the message to send
	Message Message `json:"message"`
}

// MessageEditData contains data for editing a message
type MessageEditData struct {
	// MessageID is the ID of the message to edit
	MessageID string `json:"message_id"`

	// NewContent contains the updated content
	NewContent interface{} `json:"new_content"`
}

// MessageDeleteData contains data for deleting a message
type MessageDeleteData struct {
	// MessageID is the ID of the message to delete
	MessageID string `json:"message_id"`
}

// MessageReactData contains data for reacting to a message
type MessageReactData struct {
	// MessageID is the ID of the message to react to
	MessageID string `json:"message_id"`

	// Reaction contains the reaction details
	Reaction Reaction `json:"reaction"`
}

// UserPresenceData contains data for user presence updates
type UserPresenceData struct {
	// UserID is the ID of the user
	UserID string `json:"user_id"`

	// Status is the new status of the user
	Status string `json:"status"`
}

// ActionMessage represents a complete message payload including metadata,
// channel information and the action being performed
type ActionMessage struct {
	// Metadata contains version and server node information
	Metadata Metadata `json:"metadata"`

	// Channel contains information about the messaging channel
	Channel Channel `json:"channel"`

	// Action contains the action being performed
	Action Action `json:"action"`
}
