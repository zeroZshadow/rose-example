package main

import (
	"github.com/golang/protobuf/proto"
	"github.com/zeroZshadow/rose"
	"github.com/zeroZshadow/rose-example/gameserver/client"
	"github.com/zeroZshadow/rose-example/gameserver/node"
	"github.com/zeroZshadow/rose-example/gameserver/room"
	"github.com/zeroZshadow/rose-example/messages/pb"
)

// SetupMessageHandlers Fill the message map for the client
func SetupMessageHandlers() {
	client.MessageMap[pb.MessageType_CreateRoom] = handleRoomRequest
	client.MessageMap[pb.MessageType_JoinRoom] = handleRoomRequest
}

func handleRoomRequest(user *client.User, messageType pb.MessageType, message []byte) error {
	// Unmarshal package into structure
	input := &pb.RoomRequest{}
	err := proto.Unmarshal(message, input)
	if err != nil {
		sendRoomResponse(user, messageType, false, rose.RoomID(input.Id))
		return err
	}

	roomID := rose.RoomID(input.Id)

	// Does the user already have a room? fail!
	if user.Room != nil {
		log.Warning("User already in a room")
		sendRoomResponse(user, messageType, false, roomID)
		user.Disconnect()
		return nil
	}

	// Verify authentication
	userID, err := node.Instance.VerifyAuthentication(roomID, input.Authtoken)
	if err != nil {
		log.Warningf("Invalid authentication token %s", err)
		sendRoomResponse(user, messageType, false, roomID)
		user.Disconnect()
		return nil
	}

	// Since the request is valid, we can use this to automatically login the userID
	user.ID = userID

	var result bool
	switch messageType {
	case pb.MessageType_CreateRoom:
		result = createRoom(user, messageType, roomID)
	case pb.MessageType_JoinRoom:
		result = joinRoom(user, messageType, roomID)
	}

	// Response
	sendRoomResponse(user, messageType, result, roomID)

	return nil
}

func createRoom(user *client.User, messageType pb.MessageType, roomID rose.RoomID) bool {
	// Create a new room
	roomfront := rose.RoomLobby.NewRoom(roomID, room.New)
	if roomfront == nil {
		log.Errorf("Failed to create room %d", roomID)
		return false
	}

	// Join the freshly created room
	if err := rose.RoomLobby.JoinRoom(roomID, user); err != nil {
		log.Errorf("Failed to join created room %d", roomID)
		return false
	}

	return true
}

func joinRoom(user *client.User, messageType pb.MessageType, roomID rose.RoomID) bool {
	// Join existing room
	if err := rose.RoomLobby.JoinRoom(roomID, user); err != nil {
		log.Errorf("Failed to join room %d", roomID)
		return false
	}

	return true
}

func sendRoomResponse(user *client.User, messageType pb.MessageType, success bool, roomID rose.RoomID) {
	// Create response
	response := &pb.RoomResponse{
		Success: success,
		Id:      uint64(roomID),
	}

	// Send response
	user.SendMessage(rose.MessageType(messageType), response)
}
