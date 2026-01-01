package proxmox

import (
	"raznar.id/proxmox-traffic-monitor/pkg/api/proxmox/base"
)


type ProxmoxAPI struct {
	base.BaseAPI
}

func New() ProxmoxAPI {
	return ProxmoxAPI{}
}
