// Package registry manages worker node authentication and health
package registry

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"strconv"
	"sync"
	"time"
)

type NodeStatus string
type NodeID string

const (
	NodeHealthy NodeStatus = "healthy"
	NodePending NodeStatus = "pending"
)

type WorkerConfig struct {
	ID     string `yaml:"id"`
	PubKey string `yaml:"pubkey"`
}

// heartbeat every 5s
// timeout after 15s
type Node struct {
	ID               NodeID
	MissedHeartBeats int
	Status           NodeStatus
	LastHeartBeat    time.Time
}

type Logger interface {
	Infoln(msg string)
	Warnln(msg string)
	Successln(msg string)
	Errorln(msg string)
}

type Registry struct {
	nodeRegistered <-chan NodeID
	nodeHeartBeat  <-chan NodeID
	nodeIdle       chan NodeID
	nodeDead       chan NodeID
	errorChan      chan error

	logger Logger

	pubKeyMutex      sync.RWMutex
	serverPublicKeys map[string]*rsa.PublicKey

	workerNodeMutex sync.RWMutex
	workerNodes     map[NodeID]*Node
}

func NewRegistry(l Logger, nodeHeartBeat <-chan NodeID) *Registry {
	return &Registry{
		nodeHeartBeat:    nodeHeartBeat,
		nodeIdle:         make(chan NodeID, 100),
		nodeDead:         make(chan NodeID, 100),
		logger:           l,
		pubKeyMutex:      sync.RWMutex{},
		serverPublicKeys: make(map[string]*rsa.PublicKey),
		workerNodeMutex:  sync.RWMutex{},
		workerNodes:      make(map[NodeID]*Node),
	}
}

func (r *Registry) GetIdleNodesChan() <-chan NodeID {
	return r.nodeIdle
}

func (r *Registry) GetDeadNodesChan() <-chan NodeID {
	return r.nodeDead
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

func (r *Registry) Start(ctx context.Context) {
	r.logger.Infoln("starting registry")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			r.logger.Infoln("shutting down worker heartbeat loop")
			return
		case nodeID := <-r.nodeHeartBeat:
			node, ok := r.workerNodes[nodeID]
			if !ok {
				r.errorChan <- errors.New("node (ID: " + string(node.ID) + ") is not registered")
			} else {
				node.LastHeartBeat = time.Now()
				node.Status = NodeHealthy
			}
		case <-ticker.C:
			for nodeID, node := range r.workerNodes {
				if time.Since(node.LastHeartBeat) > 6*time.Second {
					if node.MissedHeartBeats < 4 {
						node.MissedHeartBeats++
						node.Status = NodePending
						r.logger.Infoln("node (ID: " + string(node.ID) + ") is pending, missed heartbeats: " + strconv.Itoa(node.MissedHeartBeats))
					} else {
						r.logger.Warnln("node (ID: " + string(node.ID) + ") is dead")
						r.nodeDead <- nodeID
						delete(r.workerNodes, nodeID)
					}
				}
			}
		}
	}
}
