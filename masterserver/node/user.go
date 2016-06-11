package node

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/op/go-logging"
	"github.com/zeroZshadow/rose"
	"github.com/zeroZshadow/rose-example/masterserver/lobby"
	"github.com/zeroZshadow/rose-example/messages/pb"
)

type userMessageHandler func(*User, pb.MessageType, []byte)

// MessageMap Map of messageType handlers
var messageMap = make(map[pb.MessageType]userMessageHandler)
var log = logging.MustGetLogger("global")

// User is a connected game server
type User struct {
	// Framework things
	*rose.UserBase

	// Custom data
	Address     string
	Region      string
	Development bool
	RoomCount   int
	RoomMax     int
	CipherKey   []byte
}

// HandlePacket implements User.HandlePacket
func (user *User) HandlePacket(msgType rose.MessageType, message []byte) {
	//Convert to pb
	messageType := pb.MessageType(msgType)

	// Find handler for message type, run if available
	if handler, ok := messageMap[messageType]; ok {
		handler(user, messageType, message)
		return
	}

	// No handler found
	log.Warningf("Unhandled node message %d!", messageType)
}

// OnDisconnect implements User.OnDisconnect
func (user *User) OnDisconnect(err error) {
	// Remove us from the list of active nodes
	Cluster.RemoveNode(user)
	lobby.RemoveRoomsFromNode(user)
}

// OnConnect implements User.OnConnect
func (user *User) OnConnect() {
	// Register new cluster in DB
	id := Cluster.AddNode(user)
	// Change to string
	user.ID = rose.UserID(id)
}

// Encrypt Encrypt given data with the node's cipher
func (user *User) Encrypt(data []byte) ([]byte, error) {
	// Create cipher block
	block, err := aes.NewCipher(user.CipherKey)
	if err != nil {
		return nil, err
	}

	// Allocate block for cypher text
	cipherText := make([]byte, aes.BlockSize+len(data))

	// Generate IV
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Encrypt
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], data)

	return cipherText, nil
}

// New Create new node.User
func New(pump *rose.MessagePump) rose.User {
	return &User{
		UserBase: rose.NewUserBase(pump),
	}
}
