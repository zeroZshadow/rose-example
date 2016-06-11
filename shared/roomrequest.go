package shared

import "github.com/zeroZshadow/rose"

// RoomRequest authentication block for room requests
type RoomRequest struct {
	UserID    rose.UserID
	RoomID    rose.RoomID
	Timestamp int64
}
