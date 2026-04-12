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
)

type NodeStatus string
type NodeID string

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
	ID            NodeID
	Status        NodeStatus
	JobActive     bool
	LastHeartBeat time.Time
}

type Heatbeat struct {
	ID string `json:"id"`
}

type Logger interface {
	Infoln(msg string)
	Warnln(msg string)
	Successln(msg string)
	Errorln(msg string)
}

type Registry struct {
	logger Logger

	pubKeyMutex      sync.RWMutex
	serverPublicKeys map[string]*rsa.PublicKey

	workerNodeMutex sync.RWMutex
	workerNodes     map[NodeID]*Node

	deadNodeMutex sync.RWMutex
	deadNodes     []NodeID
}

func NewRegistry(l Logger) *Registry {
	return &Registry{
		logger:           l,
		pubKeyMutex:      sync.RWMutex{},
		serverPublicKeys: make(map[string]*rsa.PublicKey),
		workerNodeMutex:  sync.RWMutex{},
		workerNodes:      make(map[NodeID]*Node),
		deadNodeMutex:    sync.RWMutex{},
		deadNodes:        []NodeID{},
	}
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

func (r *Registry) InitRegistry(wc []WorkerConfig) {
	r.logger.Infoln("setting up registry")

	r.pubKeyMutex.Lock()
	defer r.pubKeyMutex.Unlock()

	for _, val := range wc {
		pub, err := parseRSAPublicKeyFromFile(val.PubKey)
		if err != nil {
			r.logger.Errorln("failed loading public key for worker " + val.ID + ": " + err.Error())
			continue
		}

		r.serverPublicKeys[val.ID] = pub
	}
}

func (r *Registry) RegisterNode(node *Node) error {
	r.workerNodeMutex.Lock()
	defer r.workerNodeMutex.Unlock()

	if _, ok := r.workerNodes[node.ID]; ok {
		return errors.New("worker already registered")
	}

	r.workerNodes[node.ID] = node
	return nil
}

func (r *Registry) GetFreeNode() (NodeID, error) {
	r.workerNodeMutex.RLock()
	for _, n := range r.workerNodes {
		if !n.JobActive {
			return n.ID, nil
		}
	}
	r.workerNodeMutex.RUnlock()
	return "", errors.New("no available node")
}

func (r *Registry) RegisterHeatbeat(nodeID NodeID) error {
	r.workerNodeMutex.Lock()
	defer r.workerNodeMutex.Unlock()
	node, ok := r.workerNodes[nodeID]
	if !ok {
		return errors.New("worker does not exist")
	}

	if !r.IsNodeDead(node.ID) {
		r.logger.Successln("node (ID: " + string(nodeID) + ") has come alive again")
	}
	node.LastHeartBeat = time.Now()
	node.Status = NodeHealthy
	return nil
}

func (r *Registry) CheckHeartbeats(ctx context.Context) {
	r.logger.Infoln("starting worker heartbeat loop")
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			r.logger.Infoln("shutting down worker heartbeat loop")
			return
		case <-ticker.C:
			r.workerNodeMutex.Lock()
			nodesToRemove := []NodeID{}
			for _, node := range r.workerNodes {
				if time.Since(node.LastHeartBeat) > 6*time.Second {
					switch node.Status {
					case NodeHealthy:
						node.Status = NodePending
						r.logger.Infoln("node (ID: " + string(node.ID) + ") has missed a heartbeat")
					case NodePending:
						node.Status = NodeDead
					case NodeDead:
						nodesToRemove = append(nodesToRemove, node.ID)
					}
				}
			}
			for _, ID := range nodesToRemove {
				r.logger.Warnln("node (ID: " + string(ID) + ") has been removed")
				delete(r.workerNodes, ID)
			}
			r.workerNodeMutex.Unlock()

			r.deadNodeMutex.Lock()
			r.deadNodes = append(r.deadNodes, nodesToRemove...)
			r.deadNodeMutex.Unlock()

			time.Sleep(3 * time.Second)
		}
	}
}

func (r *Registry) IsNodeDead(ID NodeID) bool {
	r.deadNodeMutex.RLock()
	defer r.deadNodeMutex.RUnlock()
	return slices.Contains(r.deadNodes, ID)
}
