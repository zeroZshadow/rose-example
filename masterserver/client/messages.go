package client

import (
	"encoding/json"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/golang/protobuf/proto"
	"github.com/zeroZshadow/rose"
	"github.com/zeroZshadow/rose-example/masterserver/lobby"
	"github.com/zeroZshadow/rose-example/masterserver/node"
	"github.com/zeroZshadow/rose-example/messages/pb"
	"github.com/zeroZshadow/rose-example/shared"
)

var snowflakeNode *snowflake.Node

func init() {
	// I am a terrible person that ignores the error (because I know it will never happen)
	snowflakeNode, _ = snowflake.NewNode(1)
}

// SetupMessageHandlers Fill the message map for the client
func SetupMessageHandlers() {
	messageMap[pb.MessageType_CreateRoom] = handleCreateRoomRequest
	messageMap[pb.MessageType_JoinRoom] = handleJoinRoomRequest
	messageMap[pb.MessageType_ListRooms] = handleListRoomsRequest
}

func handleCreateRoomRequest(user *User, messageType pb.MessageType, message []byte) error {
	responseType := pb.MessageType_CreateRoom

	input := &pb.CreateRoomRequest{}
	err := proto.Unmarshal(message, input)
	if err != nil {
		sendRoomResponse(user, responseType, false, 0, "", nil)
		return err
	}

	// Default packet values
	roomID := rose.RoomID(snowflakeNode.Generate())

	// Find best node to put the room on
	bestNode := node.Cluster.GetBestForRegion(input.Region)

	// Fail if we didn't find a node
	if bestNode == nil {
		log.Errorf("No nodes found for region %s", input.Region)
		sendRoomResponse(user, responseType, false, roomID, "", nil)
		return nil
	}

	// Generate data for response
	roomrequest := shared.RoomRequest{
		UserID:    user.ID,
		RoomID:    roomID,
		Timestamp: time.Now().UTC().UnixNano(),
	}

	// Marshall to json so it is ready to be encrypted
	data, err := json.Marshal(roomrequest)
	if err != nil {
		log.Error("Failed to marshal room request:", err)
		sendRoomResponse(user, responseType, false, roomID, "", nil)
		return nil
	}

	// Encrypt
	authtoken, err := bestNode.Encrypt(data)
	if err != nil {
		log.Error("Failed to encrypt room request:", err)
		sendRoomResponse(user, responseType, false, roomID, "", nil)
		return nil
	}

	// Send the new room info to the player
	sendRoomResponse(user, responseType, true, roomID, bestNode.Address, authtoken)

	return nil
}

func handleJoinRoomRequest(user *User, messageType pb.MessageType, message []byte) error {
	responseType := pb.MessageType_JoinRoom

	input := &pb.JoinRoomRequest{}
	err := proto.Unmarshal(message, input)
	if err != nil {
		sendRoomResponse(user, responseType, false, 0, "", nil)
		return err
	}

	roomID := rose.RoomID(input.Id)

	// Pick a node based on the request parameters
	info, ok := lobby.GetRoomInfo(roomID)
	if !ok {
		// Room wasn't found
		sendRoomResponse(user, responseType, false, roomID, "", nil)
		return nil
	}

	// Generate data for response
	server, ok := info.Server.(*node.User)
	if !ok {
		// Server was empty wasn't found
		log.Error("Server for requestion room is nil.")
		sendRoomResponse(user, responseType, false, roomID, "", nil)
		return nil
	}

	address := server.Address

	roomrequest := shared.RoomRequest{
		UserID:    user.ID,
		RoomID:    roomID,
		Timestamp: time.Now().UTC().UnixNano(),
	}

	// Encrypt the room request using the node's public key
	data, err := json.Marshal(roomrequest)
	if err != nil {
		log.Error("Failed to marshal room request.")
		sendRoomResponse(user, responseType, false, roomID, "", nil)
		return nil
	}

	// Encrypt
	authtoken, err := server.Encrypt(data)
	if err != nil {
		log.Error("Failed to encrypt room request.")
		sendRoomResponse(user, responseType, false, roomID, "", nil)
		return nil
	}

	sendRoomResponse(user, responseType, true, roomID, address, authtoken)

	return nil
}

func sendRoomResponse(user *User, messageType pb.MessageType, success bool, roomID rose.RoomID, address string, authtoken []byte) {
	// Create response
	response := &pb.CreateRoomResponse{
		Success:   success,
		Id:        uint64(roomID),
		Address:   address,
		Authtoken: authtoken,
	}

	// Send response
	user.SendMessage(rose.MessageType(messageType), response)
}

func handleListRoomsRequest(user *User, messageType pb.MessageType, message []byte) error {
	input := &pb.ListRoomsRequest{}
	err := proto.Unmarshal(message, input)
	if err != nil {
		return err
	}

	// Get all rooms for region
	rooms := lobby.GetAllRooms(input.Region)

	// Response
	response := &pb.ListRoomsResponse{
		Region: input.Region,
		Rooms:  rooms,
	}

	// Send response and check for errors
	user.SendMessage(rose.MessageType(pb.MessageType_ListRooms), response)

	return nil
}
