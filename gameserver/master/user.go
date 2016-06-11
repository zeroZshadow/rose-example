package master

import (
	"github.com/op/go-logging"
	"github.com/zeroZshadow/rose"
	"github.com/zeroZshadow/rose-example/gameserver/node"
)

var log = logging.MustGetLogger("global")

// User Connection to master server
type User struct {
	// Framework things
	*rose.UserBase
}

// HandlePacket Implements rose.User.HandlePacket
func (user *User) HandlePacket(messageType rose.MessageType, message []byte) {

}

// OnDisconnect Implements rose.User.OnDisconnect
func (user *User) OnDisconnect(err error) {
	node.Instance.Lock()
	defer node.Instance.Unlock()

	node.Instance.Master = nil
	log.Warning("Lost connection to master, retrying.")
}

// OnConnect Implements rose.User.OnConnect
func (user *User) OnConnect() {

}

// New create a new master user
func New(pump *rose.MessagePump) rose.User {
	return &User{
		UserBase: rose.NewUserBase(pump),
	}
}
