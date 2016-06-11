package main

import (
	"github.com/golang/protobuf/proto"
	"github.com/zeroZshadow/rose"
	"github.com/zeroZshadow/rose-example/gameserver/client"
	"github.com/zeroZshadow/rose-example/gameserver/lobby"
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

	var room *rose.RoomFront

	switch messageType {
	case pb.MessageType_CreateRoom:
		room = createRoom(user, messageType, roomID)
	case pb.MessageType_JoinRoom:
		room = joinRoom(user, messageType, roomID)
	}

	// Have the creator set the room if we have one
	user.Room = room

	// Response
	// TODO could be a bit risky.
	sendRoomResponse(user, messageType, room != nil, roomID)

	// Adding the User also sends the room info to the master
	if room != nil {
		room.QueueAddUser(user)
	}

	return nil
}

func createRoom(user *client.User, messageType pb.MessageType, roomID rose.RoomID) *rose.RoomFront {
	// Create a new room
	roomfront := lobby.Instance.NewRoom(roomID, room.New)
	if roomfront == nil {
		log.Errorf("Failed to create room %d", roomID)
		return nil
	}

	// Run the room first so any calls to the room wont block
	roomfront.Start()

	return roomfront
}

func joinRoom(user *client.User, messageType pb.MessageType, roomID rose.RoomID) *rose.RoomFront {
	// Join existing room
	roomfront := lobby.Instance.GetRoom(roomID)
	if roomfront == nil {
		log.Warningf("Failed to get room %d", roomID)
	}

	return roomfront
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
