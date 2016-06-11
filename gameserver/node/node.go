package node

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/op/go-logging"

	"github.com/zeroZshadow/rose"
	"github.com/zeroZshadow/rose-example/messages/pb"
	"github.com/zeroZshadow/rose-example/shared"
)

const (
	keySize = 24
)

var (
	// Instance a global node instance
	Instance          *Node
	masterConstructor rose.UserConstructor
	log               = logging.MustGetLogger("global")
)

// Node structure represends the connection to the master server
type Node struct {
	Master    rose.User
	server    *rose.Server
	cipherkey []byte

	sync.RWMutex

	region      string
	port        uint64
	address     string
	retryTicker *time.Ticker
	retryQuit   chan struct{}
}

// Instantiate create a new global node instance
func Instantiate(server *rose.Server, region string, address string, port uint64, constructor rose.UserConstructor) {
	masterConstructor = constructor
	Instance = new(server, region, address, port)
}

func new(server *rose.Server, region string, address string, port uint64) *Node {
	return &Node{
		server:      server,
		address:     address,
		region:      region,
		port:        port,
		retryTicker: time.NewTicker(10 * time.Second),
		retryQuit:   make(chan struct{}),
	}
}

// Start register the node tot he master if connected
func (node *Node) Start() {
	if node.Master != nil {
		return
	}

	// Register the node
	node.register(node.region, node.port)

	// Start retry loop
	go func() {
		for {
			select {
			case <-node.retryTicker.C:
				// make sure we still exist
				if node == nil {
					node.retryTicker.Stop()
					log.Info("Stopped reconnecting.")
					return
				}
				// Register the node
				node.register(node.region, node.port)
			case <-node.retryQuit:
				node.retryTicker.Stop()
				log.Info("Stopped reconnecting.")
				return
			}
		}
	}()
}

// Stop stop the node from connecting to the master
func (node *Node) Stop() {
	node.retryTicker.Stop()
	node.retryQuit <- struct{}{}
}

func (node *Node) register(region string, port uint64) {
	// Make sure no one uses the node while we register
	node.Lock()
	defer node.Unlock()

	// Do not reregister if we're already good
	if node.Master != nil {
		return
	}

	// Attempt to connect to the master
	master, err := node.server.Connect(node.address, masterConstructor)
	if err != nil || master == nil {
		// Failed to connect
		log.Warningf("Failed to connect to master: %s", node.address)
		return
	}
	log.Noticef("Connected to master: %s", node.address)

	// Save the master on the node
	node.Master = master

	// Since we're now connected, register ourselfs
	//Generate random key pass
	randomKey := make([]byte, keySize)
	_, err = rand.Read(randomKey)
	if err != nil {
		log.Fatalf("Unable to generate key %s", err)
	}
	node.cipherkey = randomKey

	// Get external address
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}

	// Find the non loopback address, or fallback to localhost
	addressString := "localhost"
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsLinkLocalUnicast() {
			if ipnet.IP.To4() != nil {
				addressString = ipnet.IP.String()
				break
			}

		}
	}

	// Attach port
	addressString = fmt.Sprintf("%s:%d", addressString, port)

	// Registration
	response := &pb.RegisterNodeRequest{
		Region:  region,
		Cipher:  randomKey,
		Address: addressString,
	}

	// Send registration
	node.Master.SendMessage(rose.MessageType(pb.MessageType_RegisterNode), response)

	log.Noticef("Registered game node with ip: %s.", addressString)
}

// VerifyAuthentication verify that the auth block from the user is correct, return the user id inside
func (node *Node) VerifyAuthentication(roomID rose.RoomID, auth []byte) (rose.UserID, error) {
	// Generate block
	block, err := aes.NewCipher(node.cipherkey)
	if err != nil {
		return 0, err
	}
	if len(auth) < aes.BlockSize {
		return 0, errors.New("ciphertext too short")
	}

	// Decrypt authentication data
	iv := auth[:aes.BlockSize]
	data := auth[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(data, data)

	// Unmarshal the decrypted data into a struct
	request := &shared.RoomRequest{}
	err = json.Unmarshal(data, request)
	if err != nil {
		return 0, err
	}

	// If the data does not match, fail the verification
	if roomID != request.RoomID {
		return 0, errors.New("Wrong roomid!")
	}

	// TODO Something with the timestamp

	return request.UserID, nil
}
