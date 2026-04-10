// Package registry
package registry

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/ljlericson/TaskForge/internal/logging"
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
	Status        NodeStatus
	JobActive     bool
	LastHeartBeat time.Time
}

type Heatbeat struct {
	ID          string `json:"id"`
	JobProgress uint8  `json:"jobProgress"`
	JobActive   bool   `json:"jobActive"`
}

type registryState struct {
	pubKeyMutex      sync.RWMutex
	serverPublicKeys map[string]*rsa.PublicKey

	workerNodeMutex sync.RWMutex
	workerNodes     map[string]*Node

	deadNodeMutex sync.RWMutex
	deadNodes     []string
}

var registryInstance *registryState = &registryState{
	pubKeyMutex:      sync.RWMutex{},
	serverPublicKeys: make(map[string]*rsa.PublicKey),
	workerNodeMutex:  sync.RWMutex{},
	workerNodes:      make(map[string]*Node),
	deadNodeMutex:    sync.RWMutex{},
	deadNodes:        []string{},
}

func parseRSAPublicKeyFromFile(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid pem")
	}

	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pub, ok := pubAny.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("is not a rsa public pem")
	}

	return pub, nil
}

func InitRegistry(wc []WorkerConfig) {
	logging.Infoln("setting up registry")

	registryInstance.pubKeyMutex.Lock()
	defer registryInstance.pubKeyMutex.Unlock()

	for _, val := range wc {
		pub, err := parseRSAPublicKeyFromFile(val.PubKey)
		if err != nil {
			logging.Errorln("failed loading public key for worker " + val.ID + ": " + err.Error())
			continue
		}

		registryInstance.serverPublicKeys[val.ID] = pub
	}
}

func RegisterNode(node *Node) error {
	registryInstance.workerNodeMutex.Lock()
	defer registryInstance.workerNodeMutex.Unlock()

	if _, ok := registryInstance.workerNodes[node.ID]; ok {
		return errors.New("worker already registered")
	}

	registryInstance.workerNodes[node.ID] = node
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
func RegisterHeatbeat(heartbeat Heatbeat) error {
	registryInstance.workerNodeMutex.Lock()
	defer registryInstance.workerNodeMutex.Unlock()
	node, ok := registryInstance.workerNodes[heartbeat.ID]
	if !ok {
		return errors.New("worker does not exist")
	}

	if !IsNodeDead(node.ID) {
		node.LastHeartBeat = time.Now()
		node.JobActive = heartbeat.JobActive
		node.Status = NodeHealthy
	}
	return nil
}

func CheckHeartbeats(ctx context.Context) {
	logging.Infoln("starting worker heartbeat loop")
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			logging.Infoln("shutting down worker heartbeat loop")
			return
		case <-ticker.C:
			registryInstance.workerNodeMutex.Lock()
			nodesToRemove := []string{}
			for _, node := range registryInstance.workerNodes {
				if time.Since(node.LastHeartBeat) > 6*time.Second {
					switch node.Status {
					case NodeHealthy:
						node.Status = NodePending
						logging.Infoln("node (ID: " + node.ID + ") has missed a heartbeat")
					case NodePending:
						node.Status = NodeDead
					case NodeDead:
						nodesToRemove = append(nodesToRemove, node.ID)
					}
				}
			}
			for _, ID := range nodesToRemove {
				logging.Warnln("node (ID: " + ID + ") has been removed")
				delete(registryInstance.workerNodes, ID)
			}
			registryInstance.workerNodeMutex.Unlock()

			registryInstance.deadNodeMutex.Lock()
			registryInstance.deadNodes = append(registryInstance.deadNodes, nodesToRemove...)
			registryInstance.deadNodeMutex.Unlock()

			time.Sleep(3 * time.Second)
		}
	}
}

func IsNodeDead(ID string) bool {
	registryInstance.deadNodeMutex.RLock()
	defer registryInstance.deadNodeMutex.RUnlock()
	return slices.Contains(registryInstance.deadNodes, ID)
}
