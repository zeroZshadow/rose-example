package room

import (
	"github.com/zeroZshadow/rose"
	"github.com/zeroZshadow/rose-example/gameserver/node"
	"github.com/zeroZshadow/rose-example/messages/pb"
)

func (room *Room) generateRoomInfo() *pb.RoomInfo {
	info := &pb.RoomInfo{
		Id: uint64(room.ID),
		// TODO Fill
	}

	return info
}

func (room *Room) updateMasterInfo(removed bool) {
	node.Instance.RLock()
	defer node.Instance.RUnlock()

	if node.Instance.Master == nil {
		return
	}

	// Inform master about the updated room
	roominfo := &pb.UpdateRoomRequest{
		Room:   room.generateRoomInfo(),
		Remove: removed,
	}

	// Send room info to master
	node.Instance.Master.SendMessage(rose.MessageType(pb.MessageType_UpdateRoom), roominfo)
}
