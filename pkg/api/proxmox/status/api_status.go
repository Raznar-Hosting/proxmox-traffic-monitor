package status

import (
	"raznar.id/proxmox-traffic-monitor/pkg/api/proxmox/base"
)

type ProxmoxStatusAPI struct {
	nodeID string
	base   base.BaseAPI
	// temporary (change to api later)
}

func New(base base.BaseAPI, nodeID string) (api ProxmoxStatusAPI) {
	api.nodeID = nodeID
	api.base = base

	return
}
