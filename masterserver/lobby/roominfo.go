package lobby

import "github.com/zeroZshadow/rose"

// RoomInfo representing a room on a node
type RoomInfo struct {
	ID          rose.RoomID
	Name        string
	PlayerCount int
	PlayerMax   int
	State       int

	Server rose.User
}
