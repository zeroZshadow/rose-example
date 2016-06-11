package client

import (
	"github.com/op/go-logging"
	"github.com/zeroZshadow/rose"
	"github.com/zeroZshadow/rose-example/messages/pb"
)

type userMessageHandler func(*User, pb.MessageType, []byte) error

// MessageMap Map of messageType handlers
var MessageMap = make(map[pb.MessageType]userMessageHandler)
var log = logging.MustGetLogger("global")

// User is a game user
type User struct {
	// Framework things
	*rose.UserBase

	Room *rose.RoomFront
}

// HandlePacket sends the recieved packet data to HandleUserPacket
func (user *User) HandlePacket(msgType rose.MessageType, message []byte) {
	messageType := pb.MessageType(msgType)

	// Find handler for message type, run if available
	if handler, ok := MessageMap[messageType]; ok {
		err := handler(user, messageType, message)
		if err != nil {
			log.Errorf("unmarshaling error: %s\n%v", err, message)
		}
		return
	}

	// Try to handle the message in the user's room instead
	if user.Room == nil {
		log.Errorf("Message of type %d arrived from a user without a room.", messageType)
		return
	}

	// Queue message in room
	user.Room.PushMessage(user, msgType, message)
}

// OnDisconnect removes the user from any connected rooms
func (user *User) OnDisconnect(err error) {
	if user.Room != nil {
		user.Room.QueueRemoveUser(user)
	}
}

// OnConnect runs whever a new user connects
func (user *User) OnConnect() {
	// Start timeout timer for login?
}

// New create new client
func New(pump *rose.MessagePump) rose.User {
	user := &User{
		UserBase: rose.NewUserBase(pump),
	}

	return user
}
