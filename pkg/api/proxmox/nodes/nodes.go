package nodes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ProxmoxNodeData represents a node's data in Proxmox.
type ProxmoxNodeData struct {
	MaxCPU         int     `json:"maxcpu"`
	Status         string  `json:"status"`
	SSLFingerprint string  `json:"ssl_fingerprint"`
	Node           string  `json:"node"`
	Uptime         int64   `json:"uptime"`
	MaxMem         int64   `json:"maxmem"`
	Type           string  `json:"type"`
	CPU            float64 `json:"cpu"`
	MaxDisk        int64   `json:"maxdisk"`
	Level          string  `json:"level"`
	Mem            int64   `json:"mem"`
	ID             string  `json:"id"`
	Disk           int64   `json:"disk"`
}

func (data ProxmoxNodeData) GetId() string {
	return strings.ReplaceAll(data.ID,"node/", "")
}



// GetNodes retrieves the list of nodes from the Proxmox API.
func (api ProxmoxNodesAPI) GetNodes() (nodesData []ProxmoxNodeData, err error) {
	baseRes, err := api.base.NewRequest("/nodes", http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := baseRes.Execute()
	if err != nil {
		return nil, fmt.Errorf("request execution failed: %w", err)
	}


	if err = json.Unmarshal(res, &nodesData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return
}

// GetFirstNode retrieves the first node from the Proxmox API.
func (api ProxmoxNodesAPI) GetFirstNode() (nodeData *ProxmoxNodeData, err error) {
	nodesData, err := api.GetNodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	if len(nodesData) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	nodeData = &nodesData[0]
	return
}