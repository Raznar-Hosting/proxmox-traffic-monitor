package qemu

import (
	"bytes"
	"encoding/json"
	"fmt"

	"raznar.id/proxmox-traffic-monitor/pkg/api/interfaces/interface_proxmox/interface_qemu"
)

func (api ProxmoxQEMUAPI) GetConfig(id int64) (data interface_qemu.VMConfig, err error) {
	endpoint := fmt.Sprintf("/nodes/%s/qemu/%d/config", api.nodeID, id)
	baseRes, err := api.base.NewRequest(endpoint, "GET", nil)
	if err != nil {
		return
	}

	res, err := baseRes.Execute()
	if err != nil {
		return
	}

	err = json.Unmarshal(res, &data)
	if err != nil {
		return
	}

	return
}

func (api ProxmoxQEMUAPI) UpdateConfig(id int64, config interface_qemu.VMConfig) (err error) {
	endpoint := fmt.Sprintf("/nodes/%s/qemu/%d/config", api.nodeID, id)

	bodyData, err := json.Marshal(config)
	if err != nil {
		return
	}

	baseRes, err := api.base.NewRequest(endpoint, "PUT", bytes.NewBuffer(bodyData))
	if err != nil {
		return
	}

	_, err = baseRes.Execute()
	return
}
