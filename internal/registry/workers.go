// Package registry
package registry

import (
	"context"
	"crypto/ed25519"
	"errors"
	"sync"
	"time"
)

type NodeStatus string

const (
	NodeHealthy NodeStatus = "healthy"
	NodePending NodeStatus = "pending"
	NodeDead    NodeStatus = "dead"
)

type WorkerConfig struct {
	ID     string `yaml:"id"`
	PubKey string `yaml:"pubkey"`
}

// heartbeat every 5s
// timeout after 15s
type Node struct {
	ID            string `json:"id"`
	Address       string `json:"address"`
	Status        NodeStatus
	JobActive     bool
	LastHeartBeat time.Time
}

type Heatbeat struct {
	ID string `json:"id"`
}

type registryState struct {
	pubKeyMutex      sync.RWMutex
	serverPublicKeys map[string]ed25519.PublicKey

	workerNodeMutex sync.RWMutex
	workerNodes     map[string]*Node
}

var registryInstance *registryState = &registryState{
	pubKeyMutex:      sync.RWMutex{},
	serverPublicKeys: make(map[string]ed25519.PublicKey),
	workerNodeMutex:  sync.RWMutex{},
	workerNodes:      make(map[string]*Node),
}

func InitRegistry(wc []WorkerConfig) {
	registryInstance.pubKeyMutex.Lock()
	for _, val := range wc {
		registryInstance.serverPublicKeys[val.ID] = ed25519.PublicKey(val.PubKey)
	}
	registryInstance.pubKeyMutex.Unlock()
}

func AuthenticateWorker(workerID string, message, signature []byte) bool {
	registryInstance.pubKeyMutex.RLock()
	pub, ok := registryInstance.serverPublicKeys[workerID]
	registryInstance.pubKeyMutex.RUnlock()

	if !ok {
		return false
	}

	return ed25519.Verify(pub, message, signature)
}

func RegisterNode(node *Node) error {
	registryInstance.workerNodeMutex.RLock()
	if _, ok := registryInstance.workerNodes[node.ID]; ok {
		return errors.New("worker already registered")
	}
	registryInstance.workerNodeMutex.RUnlock()

	registryInstance.workerNodeMutex.Lock()
	registryInstance.workerNodes[node.ID] = node
	registryInstance.workerNodeMutex.Unlock()
	return nil
}

func GetFreeNode() (*Node, error) {
	registryInstance.workerNodeMutex.RLock()
	for _, n := range registryInstance.workerNodes {
		if !n.JobActive {
			return n, nil
		}
	}
	registryInstance.workerNodeMutex.RUnlock()
	return nil, errors.New("no available node")
}

// TODO, add status and progress updates to heartbeat
func RegisterHeatbeat(ID string) error {
	registryInstance.workerNodeMutex.Lock()
	defer registryInstance.workerNodeMutex.Unlock()
	node, ok := registryInstance.workerNodes[ID]
	if !ok {
		registryInstance.workerNodeMutex.Unlock()
		return errors.New("worker does not exist")
	}

	node.LastHeartBeat = time.Now()
	return nil
}

func CheckHeartbeats(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			registryInstance.workerNodeMutex.Lock()
			nodesToRemove := []string{}
			for _, node := range registryInstance.workerNodes {
				if time.Since(node.LastHeartBeat) > 6*time.Second {
					switch node.Status {
					case NodeHealthy:
						node.Status = NodePending
					case NodePending:
						node.Status = NodeDead
					case NodeDead:
						if !node.JobActive {
							nodesToRemove = append(nodesToRemove, node.ID)
						}
						// alert schedular to failed job
					}
				}
			}
			for _, ID := range nodesToRemove {
				delete(registryInstance.workerNodes, ID)
			}
			registryInstance.workerNodeMutex.Unlock()
			time.Sleep(3 * time.Second)
		}
	}
}
