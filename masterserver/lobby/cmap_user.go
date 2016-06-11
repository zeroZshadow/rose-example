package lobby

import (
	"encoding/json"
	"hash/fnv"
	"sync"

	"github.com/zeroZshadow/rose"
)

var shardCountUsers = 512

// A "thread" safe map of type rose.UserID:rose.User.
// To avoid lock bottlenecks this map is dived to several (shardCountUsers) map shards.

// ConcurrentUserMap collection of user map shards
type ConcurrentUserMap []*ConcurrentUserMapShard

// ConcurrentUserMapShard user map shard
type ConcurrentUserMapShard struct {
	items        map[rose.UserID]rose.User
	sync.RWMutex // Read Write mutex, guards access to internal map.
}

// NewUserMap Creates a new concurrent map.
func NewUserMap() ConcurrentUserMap {
	m := make(ConcurrentUserMap, shardCountUsers)
	for i := 0; i < shardCountUsers; i++ {
		m[i] = &ConcurrentUserMapShard{items: make(map[rose.UserID]rose.User)}
	}
	return m
}

// GetShard Returns shard under given key
func (m ConcurrentUserMap) GetShard(key rose.UserID) *ConcurrentUserMapShard {
	hasher := fnv.New32()
	hasher.Write([]byte(string(key)))
	return m[hasher.Sum32()%uint32(shardCountUsers)]
}

// Set Sets the given value under the specified key.
func (m *ConcurrentUserMap) Set(key rose.UserID, value rose.User) {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	shard.items[key] = value
}

// Get Retrieves an element from map under given key.
func (m ConcurrentUserMap) Get(key rose.UserID) (rose.User, bool) {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()

	// Get item from shard.
	val, ok := shard.items[key]
	return val, ok
}

// Count Returns the number of elements within the map.
func (m ConcurrentUserMap) Count() int {
	count := 0
	for i := 0; i < shardCountUsers; i++ {
		shard := m[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

// Has Looks up an item under specified key
func (m *ConcurrentUserMap) Has(key rose.UserID) bool {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()

	// See if element is within shard.
	_, ok := shard.items[key]
	return ok
}

// Remove Removes an element from the map.
func (m *ConcurrentUserMap) Remove(key rose.UserID) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	delete(shard.items, key)
}

// IsEmpty Checks if map is empty.
func (m *ConcurrentUserMap) IsEmpty() bool {
	return m.Count() == 0
}

// TupleUser Used by the Iter & IterBuffered functions to wrap two variables together over a channel,
type TupleUser struct {
	Key rose.UserID
	Val rose.User
}

// Iter Returns an iterator which could be used in a for range loop.
func (m ConcurrentUserMap) Iter() <-chan TupleUser {
	ch := make(chan TupleUser)
	go func() {
		// Foreach shard.
		for _, shard := range m {
			// Foreach key, value pair.
			shard.RLock()
			for key, val := range shard.items {
				ch <- TupleUser{key, val}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}

// IterBuffered Returns a buffered iterator which could be used in a for range loop.
func (m ConcurrentUserMap) IterBuffered() <-chan TupleUser {
	ch := make(chan TupleUser, m.Count())
	go func() {
		// Foreach shard.
		for _, shard := range m {
			// Foreach key, value pair.
			shard.RLock()
			for key, val := range shard.items {
				ch <- TupleUser{key, val}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}

// MarshalJSON Implement the Marshaller
func (m ConcurrentUserMap) MarshalJSON() ([]byte, error) {
	// Create a temporary map, which will hold all item spread across shards.
	tmp := make(map[rose.UserID]rose.User)

	// Insert items to temporary map.
	for item := range m.Iter() {
		tmp[item.Key] = item.Val
	}
	return json.Marshal(tmp)
}

// UnmarshalJSON Implement the Unmarshaller
func (m *ConcurrentUserMap) UnmarshalJSON(b []byte) (err error) {
	// Reverse process of Marshal.

	tmp := make(map[rose.UserID]rose.User)

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
