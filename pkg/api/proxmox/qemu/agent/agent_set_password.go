package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type QemuGuestPasswordData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (api ProxmoxAgentAPI) SetPassword(id int64, username string, password string) (err error) {
	byteData, err := json.Marshal(QemuGuestPasswordData{
		Username: username,
		Password: password,
	})
	if err != nil {
		return
	}

	endpoint := fmt.Sprintf("/nodes/%s/qemu/%d/agent/set-user-password", api.nodeID, id)
	baseRes, err := api.base.NewRequest(endpoint, http.MethodPost, bytes.NewBuffer(byteData))
	if err != nil {
		return
	}

	_, err = baseRes.Execute()
	if err != nil {
		return
	}

	return
}
