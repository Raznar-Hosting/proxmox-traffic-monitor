package agent

import (
	"raznar.id/proxmox-traffic-monitor/pkg/api/proxmox/base"
)

type ProxmoxAgentAPI struct {
	nodeID string
	base   base.BaseAPI
	// temporary (change to api later)
}

func New(base base.BaseAPI, nodeID string) (api ProxmoxAgentAPI) {
	api.nodeID = nodeID
	api.base = base

	return
}

// TODO: create cmd exec
