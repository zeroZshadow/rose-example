package client

import (
	"github.com/op/go-logging"
	"github.com/zeroZshadow/rose"
	"github.com/zeroZshadow/rose-example/masterserver/lobby"
	"github.com/zeroZshadow/rose-example/messages/pb"
)

// Logging
var log = logging.MustGetLogger("global")

type userMessageHandler func(*User, pb.MessageType, []byte) error

// MessageMap Map of messageType handlers
var messageMap = make(map[pb.MessageType]userMessageHandler)

// User Client user
type User struct {
	// Framework things
	*rose.UserBase
}

// HandlePacket implements rose.User.HandlePacket
func (user *User) HandlePacket(msgType rose.MessageType, message []byte) {
	//Convert to pb
	messageType := pb.MessageType(msgType)

	// Find handler for message type, run if available
	if handler, ok := messageMap[messageType]; ok {
		err := handler(user, messageType, message)
		if err != nil {
			log.Errorf("unmarshaling error: %s\n%v", err, message)
		}
		return
	}

	// No handler found
	log.Warningf("Unhandled client message %d!", messageType)
}

// OnDisconnect implements rose.User.OnDisconnect
func (user *User) OnDisconnect(err error) {
	lobby.RemoveUser(user.ID)
	log.Debug("A user disconnected.")
}

// OnConnect implements rose.User.OnConnect
func (user *User) OnConnect() {
	log.Debug("A user connected.")
}

// New create a new client.User
func New(pump *rose.MessagePump) rose.User {
	return &User{
		UserBase: rose.NewUserBase(pump),
	}
}
