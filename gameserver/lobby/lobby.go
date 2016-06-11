package lobby

import (
	"sync"

	"github.com/op/go-logging"
	"github.com/zeroZshadow/rose"
)

type roomConstructor func(rose.RoomID) rose.Room

// Instance global lobby instance
var Instance *Lobby
var log = logging.MustGetLogger("global")

//TODO User sharding similar to master server

// Lobby collection of rooms
type Lobby struct {
	rooms map[rose.RoomID]*rose.RoomFront
	sync.RWMutex
}

func new() *Lobby {
	return &Lobby{
		rooms: make(map[rose.RoomID]*rose.RoomFront),
	}
}

// Instantiate Create new global lobby Instance
func init() {
	Instance = new()
}

// NewRoom Create a new room and save it in the lobby
func (lobby *Lobby) NewRoom(id rose.RoomID, constructor roomConstructor) *rose.RoomFront {
	// Lock room list for writing
	lobby.Lock()

	// Only create a new room if the ID is not yet taken
	var front *rose.RoomFront
	if _, ok := lobby.rooms[id]; !ok {
		// Create the new room
		front = rose.NewRoomFront(constructor(id))
		lobby.rooms[id] = front
	} else {
		log.Warningf("Trying to create already existing room %s", id)
	}

	lobby.Unlock()

	return front
}

// RemoveRoom Remove room with given ID from the collection
func (lobby *Lobby) RemoveRoom(id rose.RoomID) {
	// Lock room list for writing
	lobby.Lock()

	delete(lobby.rooms, id)

	lobby.Unlock()
}

// GetRoom return the room with given Id
func (lobby *Lobby) GetRoom(id rose.RoomID) *rose.RoomFront {
	// Lock room list for reading
	lobby.RLock()

	// Get room from the RoomLobby
	front := lobby.rooms[id]

	lobby.RUnlock()

	return front
}
