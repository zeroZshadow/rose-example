package lobby

import (
	"encoding/json"
	"hash/fnv"
	"sync"

	"github.com/zeroZshadow/rose"
)

const shardCount = 128

// ConcurrentRoomInfoMap A "thread" safe map of type string:RoomInfo.
// To avoid lock bottlenecks this map is dived to several (shardCount) map shards.
type ConcurrentRoomInfoMap []*ConcurrentRoomInfoMapShard

// ConcurrentRoomInfoMapShard Shard for ConcurrentRoomInfoMap
type ConcurrentRoomInfoMapShard struct {
	items        map[rose.RoomID]RoomInfo
	sync.RWMutex // Read Write mutex, guards access to internal map.
}

// NewRoomInfoMap Creates a new concurrent map.
func NewRoomInfoMap() ConcurrentRoomInfoMap {
	m := make(ConcurrentRoomInfoMap, shardCount)
	for i := 0; i < shardCount; i++ {
		m[i] = &ConcurrentRoomInfoMapShard{items: make(map[rose.RoomID]RoomInfo)}
	}
	return m
}

// GetShard Returns shard under given key
func (m ConcurrentRoomInfoMap) GetShard(key rose.RoomID) *ConcurrentRoomInfoMapShard {
	hasher := fnv.New32()
	hasher.Write([]byte(string(key)))
	return m[hasher.Sum32()%uint32(shardCount)]
}

// Set Sets the given value under the specified key.
func (m *ConcurrentRoomInfoMap) Set(key rose.RoomID, value RoomInfo) {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	shard.items[key] = value
}

// Get Retrieves an element from map under given key.
func (m ConcurrentRoomInfoMap) Get(key rose.RoomID) (RoomInfo, bool) {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()

	// Get item from shard.
	val, ok := shard.items[key]
	return val, ok
}

// Count Returns the number of elements within the map.
func (m ConcurrentRoomInfoMap) Count() int {
	count := 0
	for i := 0; i < shardCount; i++ {
		shard := m[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

// Has Looks up an item under specified key
func (m *ConcurrentRoomInfoMap) Has(key rose.RoomID) bool {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()

	// See if element is within shard.
	_, ok := shard.items[key]
	return ok
}

// Remove Removes an element from the map.
func (m *ConcurrentRoomInfoMap) Remove(key rose.RoomID) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	delete(shard.items, key)
}

// IsEmpty Checks if map is empty.
func (m *ConcurrentRoomInfoMap) IsEmpty() bool {
	return m.Count() == 0
}

// TupleRoomInfo Used by the Iter & IterBuffered functions to wrap two variables together over a channel,
type TupleRoomInfo struct {
	Key rose.RoomID
	Val RoomInfo
}

// Iter Returns an iterator which could be used in a for range loop.
func (m ConcurrentRoomInfoMap) Iter() <-chan TupleRoomInfo {
	ch := make(chan TupleRoomInfo)
	go func() {
		// Foreach shard.
		for _, shard := range m {
			// Foreach key, value pair.
			shard.RLock()
			for key, val := range shard.items {
				ch <- TupleRoomInfo{key, val}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}

// IterBuffered Returns a buffered iterator which could be used in a for range loop.
func (m ConcurrentRoomInfoMap) IterBuffered() <-chan TupleRoomInfo {
	ch := make(chan TupleRoomInfo, m.Count())
	go func() {
		// Foreach shard.
		for _, shard := range m {
			// Foreach key, value pair.
			shard.RLock()
			for key, val := range shard.items {
				ch <- TupleRoomInfo{key, val}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}

// MarshalJSON Reviles ConcurrentMap "private" variables to json marshal.
func (m ConcurrentRoomInfoMap) MarshalJSON() ([]byte, error) {
	// Create a temporary map, which will hold all item spread across shards.
	tmp := make(map[rose.RoomID]RoomInfo)

	// Insert items to temporary map.
	for item := range m.Iter() {
		tmp[item.Key] = item.Val
	}
	return json.Marshal(tmp)
}

// UnmarshalJSON Unmarshals JSON into a map
func (m *ConcurrentRoomInfoMap) UnmarshalJSON(b []byte) (err error) {
	// Reverse process of Marshal.

	tmp := make(map[rose.RoomID]RoomInfo)

	// Unmarshal into a single map.
	if err := json.Unmarshal(b, &tmp); err != nil {
		return nil
	}

	// foreach key,value pair in temporary map insert into our concurrent map.
	for key, val := range tmp {
		m.Set(key, val)
	}
	return nil
}
