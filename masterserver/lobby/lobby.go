package lobby

import (
	"github.com/op/go-logging"
	"github.com/zeroZshadow/rose"
	"github.com/zeroZshadow/rose-example/messages/pb"
)

var log = logging.MustGetLogger("global")

type lobby struct {
	rooms ConcurrentRoomInfoMap
	users ConcurrentUserMap
}

var (
	instance *lobby
)

func init() {
	instance = &lobby{
		rooms: NewRoomInfoMap(),
		users: NewUserMap(),
	}
}

// GetRoomInfo get info about given room
func GetRoomInfo(id rose.RoomID) (RoomInfo, bool) {
	// Get room from the RoomLobby
	return instance.rooms.Get(id)
}

// SetRoomInfo add/set roominfo in lobby
func SetRoomInfo(roominfo RoomInfo) {
	// Add or update room info in map
	instance.rooms.Set(roominfo.ID, roominfo)
}

// RemoveRoomInfo remove room from lobby
func RemoveRoomInfo(id rose.RoomID) {
	// Remove room id from map
	instance.rooms.Remove(id)
}

// RemoveRoomsFromNode remove all rooms hosted on given node
func RemoveRoomsFromNode(node rose.User) {
	// Slightly nasty, since technically this can cause a race condition!
	// However, since we only run then when a server is down, we should not get any updates on the rooms that we are going to remove
	// Remove all rooms associated with the given server
	for pair := range instance.rooms.IterBuffered() {
		if pair.Val.Server == node {
			// Remove room id from map
			instance.rooms.Remove(pair.Key)
		}
	}
}

// GetAllRooms return all rooms for given region
func GetAllRooms(region string) []*pb.RoomInfo {
	// Get room from the RoomLobby
	rooms := make([]*pb.RoomInfo, 0, instance.rooms.Count())

	for pair := range instance.rooms.IterBuffered() {
		// Create RoomInfo to describe the room
		room := pair.Val
		info := &pb.RoomInfo{
			Id:          uint64(room.ID),
			Name:        room.Name,
			PlayerCount: int32(room.PlayerCount),
			PlayerMax:   int32(room.PlayerMax),
		}

		// Save description in list
		rooms = append(rooms, info)
	}

	return rooms
}

// SetUser Add the user to the concurrent map
func SetUser(user rose.User) {
	// Add or update user in map
	instance.users.Set(user.Base().ID, user)
}

// GetUser Grab the user for the given id from the concurrent map
func GetUser(id rose.UserID) (rose.User, bool) {
	// Remove user id from map
	user, ok := instance.users.Get(id)
	return user, ok
}

// RemoveUser Remove the user for the given id from the concurrent map
func RemoveUser(id rose.UserID) {
	// Remove user id from map
	instance.users.Remove(id)
}
