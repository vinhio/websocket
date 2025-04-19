package main

import (
	"github.com/gflydev/core"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	app := core.New()

	// Register router
	app.RegisterRouter(func(g core.IFly) {
		g.GET("/ws", NewWSHandler())
	})

	/*// Create a new instance of MessageSend
	chatMessage := data.MessageSend{}

	// Set metadata fields
	chatMessage.Metadata.Version = "1.0"
	chatMessage.Metadata.Timestamp = time.Now()
	chatMessage.Metadata.ServerNode.ID = "node-a42b7"
	chatMessage.Metadata.ServerNode.Region = "us-east-1"
	chatMessage.Metadata.ServerNode.Load = 0.75

	// Set channel fields
	chatMessage.Channel.ID = "ch-93e56f"
	chatMessage.Channel.Type = "group"
	chatMessage.Channel.CreatedAt = time.Now().Add(-24 * time.Hour) // Created 24 hours ago
	chatMessage.Channel.UpdatedAt = time.Now()

	// Add participants to the channel
	participant1 := struct {
		UserID   string    `json:"user_id"`
		Username string    `json:"username"`
		Status   string    `json:"status"`
		JoinedAt time.Time `json:"joined_at"`
	}{
		UserID:   "user-123",
		Username: "alice_dev",
		Status:   "online",
		JoinedAt: time.Now().Add(-24 * time.Hour),
	}

	participant2 := struct {
		UserID   string    `json:"user_id"`
		Username string    `json:"username"`
		Status   string    `json:"status"`
		JoinedAt time.Time `json:"joined_at"`
	}{
		UserID:   "user-456",
		Username: "bob_qa",
		Status:   "away",
		JoinedAt: time.Now().Add(-23 * time.Hour),
	}

	chatMessage.Channel.Participants = append(
		chatMessage.Channel.Participants,
		participant1,
		participant2,
	)

	// Set message fields
	chatMessage.Message = data.Message{
		ID:        "msg-78f90a",
		SenderID:  "user-123",
		Timestamp: time.Now(),
		Type:      "text",
		Content: data.ContentText{
			Text: "Hey @bob_qa can you review my pull request?",
		},
		Status:    "delivered",
		Reactions: []data.Reaction{},
	}

	// Add a reaction (using empty interface)
	reaction := data.Reaction{
		Emoji: "👍",
		Count: 2,
		Users: []string{"user-456"},
	}
	chatMessage.Message.Reactions = append(chatMessage.Message.Reactions, reaction)

	marshal, err := msgpack.Marshal(chatMessage)
	if err != nil {
		return
	}

	var item data.MessageSend
	err = msgpack.Unmarshal(marshal, &item)
	if err != nil {
		panic(err)
	}

	litter.Dump(item)*/

	app.Run()
}
