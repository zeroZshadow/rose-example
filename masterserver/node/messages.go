package node

// A node is a gameserver. This logic is for internal things, has nothing to do with the client.

import (
	"github.com/golang/protobuf/proto"
	"github.com/zeroZshadow/rose"
	"github.com/zeroZshadow/rose-example/masterserver/lobby"
	"github.com/zeroZshadow/rose-example/messages/pb"
)

// SetupMessageHandlers Fill the message map for the client
func SetupMessageHandlers() {
	messageMap[pb.MessageType_RegisterNode] = handleRegisterNode // Not to be confused with the Client's handleRegisterAccount
	messageMap[pb.MessageType_UpdateRoom] = handleUpdateRoom
}

func handleRegisterNode(user *User, messageType pb.MessageType, message []byte) {
	input := &pb.RegisterNodeRequest{}
	err := proto.Unmarshal(message, input)
	if err != nil {
		log.Errorf("unmarshaling error: %s", err)
		return
	}

	// Save registration data
	user.Region = input.Region
	user.CipherKey = input.Cipher
	user.Address = input.Address

	log.Noticef("Node %d serving at %s for region %s", user.ID, user.Address, user.Region)
}

func handleUpdateRoom(user *User, messageType pb.MessageType, message []byte) {
	input := &pb.UpdateRoomRequest{}
	err := proto.Unmarshal(message, input)
	if err != nil {
		log.Errorf("unmarshaling error: %s", err)
		return
	}

	// Check of room exists
	inputroom := input.GetRoom()
	room, ok := lobby.GetRoomInfo(rose.RoomID(inputroom.Id))
	if ok {
		// If the room exists, update
		// Do we have to remove the room?
		if input.Remove {
			// Remove room from lobby
			lobby.RemoveRoomInfo(rose.RoomID(inputroom.Id))
		} else {
			// Update room info
			room.Name = inputroom.Name
			room.PlayerCount = int(inputroom.PlayerCount)
			room.PlayerMax = int(inputroom.PlayerMax)
			room.State = int(inputroom.State)

			lobby.SetRoomInfo(room)
		}
	} else {
		// Else create the room
		room = lobby.RoomInfo{
			ID:          rose.RoomID(inputroom.Id),
			Name:        inputroom.Name,
			PlayerCount: int(inputroom.PlayerCount),
			PlayerMax:   int(inputroom.PlayerMax),
			State:       int(inputroom.State),
			Server:      user,
		}

		// Add room to lobby
		lobby.SetRoomInfo(room)
	}
}
