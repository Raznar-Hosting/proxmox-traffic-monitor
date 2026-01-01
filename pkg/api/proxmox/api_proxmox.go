package proxmox

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"raznar.id/proxmox-traffic-monitor/pkg/api/proxmox/base"
	"raznar.id/proxmox-traffic-monitor/pkg/api/proxmox/nodes"
	"raznar.id/proxmox-traffic-monitor/pkg/api/proxmox/qemu"
)

// NewRequest creates a new HTTP request for the Proxmox API.
func (api *ProxmoxAPI) NewRequest(path, method string, body io.Reader) (*base.BaseReq, error) {
	fullURL := fmt.Sprintf("/%s", strings.TrimPrefix(path, "/"))

	// Create the HTTP request
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	// Wrap request in BaseResp
	baseReq := &base.BaseReq{}
	baseReq.Request = req
	return baseReq, err
}

// NodesAPI returns the Nodes API interface.
func (api *ProxmoxAPI) NodesAPI() nodes.ProxmoxNodesAPI {
	return nodes.New(api)
}

// QEMUAPI returns the QEMU API interface, ensuring thread-safe initialization.
func (api *ProxmoxAPI) QEMUAPI(nodeID string) *qemu.ProxmoxQEMUAPI {
	return qemu.New(api, nodeID)
}
