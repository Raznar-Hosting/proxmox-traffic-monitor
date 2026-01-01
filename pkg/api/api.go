package api

import (
	"raznar.id/proxmox-traffic-monitor/pkg/api/proxmox"
)

type API struct {
	Proxmox *proxmox.ProxmoxAPI
}

func Proxmox() proxmox.ProxmoxAPI {
	return proxmox.New()
}

func New(proxmox *proxmox.ProxmoxAPI) API {
	return API{
		Proxmox: proxmox,
	}
}
