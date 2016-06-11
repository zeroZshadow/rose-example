package room

import (
	"github.com/golang/protobuf/proto"
	"github.com/zeroZshadow/rose"
	"github.com/zeroZshadow/rose-example/gameserver/client"
	"github.com/zeroZshadow/rose-example/messages/pb"
)

// SetupMessageHandlers handles incoming messages from the client
func SetupMessageHandlers() {
	messageMap[pb.MessageType_Chat] = handleChatMessage
}

func handleChatMessage(room *Room, user *client.User, messageType pb.MessageType, message []byte) error {
	// Unmarshal message into ChatMessage
	input := &pb.ChatMessage{}
	err := proto.Unmarshal(message, input)
	if err != nil {
		return err
	}

	// Debug print the chat message
	log.Debug(input.Message)

	// Send message to all connected users in the room
	room.Broadcast(rose.MessageType(pb.MessageType_Chat), input)

	return nil
}
