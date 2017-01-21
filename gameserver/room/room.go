package room

import (
	"time"

	"github.com/op/go-logging"
	"github.com/zeroZshadow/rose"
	"github.com/zeroZshadow/rose-example/gameserver/client"
	"github.com/zeroZshadow/rose-example/messages/pb"
)

type messageHandler func(*Room, *client.User, pb.MessageType, []byte) error

var (
	// MessageMap Map of messageType handlers
	messageMap = make(map[pb.MessageType]messageHandler)
	log        = logging.MustGetLogger("global")
	tickrate   = time.Second / 60
)

// Room will have to be concurrent, all functions altering the room
// or dealing with the room will have to be private.
// With the exception to the functions passing the events along
type Room struct {
	// Framework
	*rose.RoomBase
}

// New create a new Room
func New(id rose.RoomID) rose.Room {
	return &Room{
		RoomBase: rose.NewRoomBase(id, tickrate),
	}
}

// Base Implement Room.Base
func (room *Room) Base() *rose.RoomBase {
	return room.RoomBase
}

// Tick Implement Room.Tick
func (room *Room) Tick() {

}

// HandleMessage implements rose.Room.HandleMessage
func (room *Room) HandleMessage(user rose.User, msgType rose.MessageType, message []byte) {
	messageType := pb.MessageType(msgType)

	// Handle message according to type
	if handler, ok := messageMap[messageType]; ok {
		// Handle packet
		err := handler(room, user.(*client.User), messageType, message)
		if err != nil {
			log.Errorf("room error: %s\n%v", err, message)
		}
		return
	}

	log.Errorf("Unhandled messageType %v from %d", messageType, user.Base().ID)
}

// Cleanup implements rose.Room.Cleanup
func (room *Room) Cleanup() {
	// Inform master about the removed room
	room.updateMasterInfo(true)

	// Run base destroy
	room.RoomBase.Cleanup()
}

// AddUser is overwriting RoomBase.AddUser
func (room *Room) AddUser(user rose.User) {
	// Assert user to client
	userClient, ok := user.(*client.User)
	if !ok {
		log.Critical("Non-client user tried to join room")
		return
	}

	// Add user to the room
	room.RoomBase.AddUser(userClient)

	// TODO Tell other users I'm here
	log.Debugf("A new user joined room %d", room.ID)

	// Tell the master server about the new user
	room.updateMasterInfo(false)
}

// RemoveUser is overwriting RoomBase.RemoveUser
func (room *Room) RemoveUser(user rose.User) {
	// Assert user to client
	userClient, ok := user.(*client.User)
	if !ok {
		log.Critical("Non-client user tried to leave room")
		return
	}

	room.RoomBase.RemoveUser(userClient)

	// TODO Tell other users I've left
	log.Debugf("A user left room %d", room.ID)

	// Tell the master server that a user left
	room.updateMasterInfo(false)
}
