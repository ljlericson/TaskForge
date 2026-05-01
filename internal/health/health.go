package health

import (
	"time"

	"github.com/ljlericson/TaskForge/internal/logging"
)

type Health struct {
	schedulerCheckHealth chan chan struct{}
	registryCheckHealth  chan chan struct{}
	logger               *logging.Logger
}

func NewHealth(schedulerCheckHealth chan chan struct{}, registryCheckHealth chan chan struct{}, logger *logging.Logger) *Health {
	return &Health{
		schedulerCheckHealth: schedulerCheckHealth,
		registryCheckHealth:  registryCheckHealth,
		logger:               logger,
	}
}

func (h *Health) CheckHealth() {
	orange := "\033[33m"
	green := "\033[32m"
	reset := "\033[0m"
	clearLine := "\033[2K"
	carriage := "\r"

	check := func(name string, ch chan chan struct{}) bool {
		reply := make(chan struct{})
		print(clearLine + carriage + orange + "[--] checking " + name + " health" + reset)
		ch <- reply
		select {
		case <-reply:
			print(clearLine + carriage + green + "[OK] " + name + " healthy" + reset + "\n")
			return true
		case <-time.After(2 * time.Second):
			print(clearLine + carriage + "\033[31m" + "[--] " + name + " unresponsive!" + reset + "\n")
			return false
		}
	}

	sch := check("scheduler", h.schedulerCheckHealth)
	reg := check("registry", h.registryCheckHealth)

	if !sch {
		h.logger.Errorln("scheduler is in unresponsive, please restart the server")
	}

	if !reg {
		h.logger.Errorln("registry is unresponsive, please restart the server")
	}

	if reg && sch {
		h.logger.Successln("registry and scheduler healthy")
	}
}
