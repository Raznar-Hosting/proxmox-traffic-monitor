package status

import (
	"fmt"
	"net/http"
)

func (api ProxmoxStatusAPI) Restart(id int64) (err error) {
	endpoint := fmt.Sprintf("/nodes/%s/qemu/%d/status/reboot", api.nodeID, id)
	baseRes, err := api.base.NewRequest(endpoint, http.MethodPost, nil)
	if err != nil {
		err = fmt.Errorf("failed to create request: %w", err)
		return
	}

	_, err = baseRes.Execute()
	if err != nil {
		err = fmt.Errorf("request execution failed: %w", err)
		return
	}

	return
}
