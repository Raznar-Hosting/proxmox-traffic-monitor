package nodes

import (
	"raznar.id/proxmox-traffic-monitor/pkg/api/proxmox/base"
)

type ProxmoxNodesAPI struct {
	base base.BaseAPI
	// temporary (change to api later)
}

func New(base base.BaseAPI) (api ProxmoxNodesAPI) {
	api.base = base

	return
}
