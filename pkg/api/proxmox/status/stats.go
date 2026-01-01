package status

import (
	"encoding/json"
	"fmt"
	"net/http"

	"raznar.id/proxmox-traffic-monitor/pkg/api/interfaces/interface_proxmox/interface_qemu/interface_status"
)

func (api ProxmoxStatusAPI) Stats(id int64) (data *interface_status.Data, err error) {
	endpoint := fmt.Sprintf("/nodes/%s/qemu/%d/status/current", api.nodeID, id)
	baseRes, err := api.base.NewRequest(endpoint, http.MethodGet, nil)
	if err != nil {
		err = fmt.Errorf("failed to create request: %w", err)
		return
	}

	resBytes, err := baseRes.Execute() // Assume this now returns []byte
	if err != nil {
		err = fmt.Errorf("request execution failed: %w", err)
		return
	}

	if err = json.Unmarshal(resBytes, &data); err != nil {
		err = fmt.Errorf("failed to unmarshal response: %w", err)
		return
	}

	return
}
