package qemu

import (
	"raznar.id/proxmox-traffic-monitor/pkg/api/proxmox/base"
	"raznar.id/proxmox-traffic-monitor/pkg/api/proxmox/qemu/agent"
	"raznar.id/proxmox-traffic-monitor/pkg/api/proxmox/status"
)

type ProxmoxQEMUAPI struct {
	base   base.BaseAPI
	nodeID string
}

func New(base base.BaseAPI, nodeId string) *ProxmoxQEMUAPI {
	return &ProxmoxQEMUAPI{
		base:   base,
		nodeID: nodeId,
	}
}

// NodesAPI returns the Nodes API interface.
func (api ProxmoxQEMUAPI) StatusAPI() status.ProxmoxStatusAPI {
	return status.New(api.base, api.nodeID)
}

func (api ProxmoxQEMUAPI) AgentAPI() agent.ProxmoxAgentAPI {
	return agent.New(api.base, api.nodeID)
}

func (api ProxmoxQEMUAPI) NodeID() string {
	return api.nodeID
}
