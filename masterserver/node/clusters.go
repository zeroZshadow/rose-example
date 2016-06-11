package node

import (
	"sync"
)

// ClusterMap collection of nodes connected to master server
type ClusterMap struct {
	nodes     []*User
	idCounter uint64
	sync.RWMutex
}

// Cluster global cluster of nodes
var Cluster *ClusterMap

func init() {
	Cluster = &ClusterMap{
		nodes:     make([]*User, 0),
		idCounter: uint64(0),
	}
}

// AddNode add node to cluster
func (clusterMap *ClusterMap) AddNode(node *User) uint64 {
	// Lock nodes list for writing
	clusterMap.Lock()

	// Generate a new ID for the node
	id := clusterMap.idCounter
	clusterMap.idCounter++

	// Add node to list
	clusterMap.nodes = append(clusterMap.nodes, node)

	clusterMap.Unlock()

	return id
}

// RemoveNode remove a node from the cluster
func (clusterMap *ClusterMap) RemoveNode(node *User) {
	// Lock nodes list for writing
	clusterMap.Lock()
	defer clusterMap.Unlock()

	for i, n := range clusterMap.nodes {
		if n == node {
			// Remove node from the list
			clusterMap.nodes, clusterMap.nodes[len(clusterMap.nodes)-1] = append(clusterMap.nodes[:i], clusterMap.nodes[i+1:]...), nil
			return
		}
	}
}

// GetAllNodes return all nodes in the cluster
func (clusterMap *ClusterMap) GetAllNodes() []*User {
	// Lock nodes list for reading
	clusterMap.RLock()
	defer clusterMap.RUnlock()

	nodeCount := len(clusterMap.nodes)
	if nodeCount > 0 {
		nodes := make([]*User, 0, len(clusterMap.nodes))
		copy(nodes, clusterMap.nodes)
		return nodes
	}

	return nil
}

// GetBestForRegion return the lowest room count node for the given region
func (clusterMap *ClusterMap) GetBestForRegion(region string) *User {
	clusterMap.RLock()
	defer clusterMap.RUnlock()

	// Only succeed if we have nodes
	var bestNode *User

	// Iterate over nodes to get the matching region with the least rooms
	for _, node := range clusterMap.nodes {
		if node.Region == region {
			if bestNode == nil || node.RoomCount < bestNode.RoomCount {
				bestNode = node
			}
		}
	}

	return bestNode
}
