package qemu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"raznar.id/proxmox-traffic-monitor/pkg/api/interfaces/interface_proxmox/interface_qemu"
)

func (api ProxmoxQEMUAPI) Create(name string, macID string, memory int64, cores int64, balloon bool, networkRate int64) (vmid int64, err error) {
	servers, err := api.GetAllServers()
	if err != nil {
		return 0, fmt.Errorf("failed to get all servers: %w", err)
	}

	var vmIDS []int64
	for _, s := range servers {
		if s.Name == name {
			return 0, fmt.Errorf("a server with the same name already exists")
		}
		vmIDS = append(vmIDS, s.VMID)
	}

	vmid = generateNextAvailableVMID(100, vmIDS)

	netData := fmt.Sprintf("bridge=vmbr0,virtio=%s", macID)
	if networkRate > 0 {
		netData = fmt.Sprintf("%s,rate=%d", netData, networkRate)
	}

	reqBody := interface_qemu.CreateVMConfig{
		VMConfig: interface_qemu.VMConfig{
			SCSIHW:  "virtio-scsi-pci",
			Memory:  fmt.Sprintf("%d", memory),
			CPU:     "host", // corrected
			Cores:   cores,
			Net0:    netData,
			Name:    name,
			Agent:   "enabled=1",
			Serial0: "socket",
		},
		VMID: vmid,
	}

	if !balloon {
		v := 0
		reqBody.Balloon = &v
	} else {
		mem, _ := strconv.Atoi(reqBody.Memory)
		reqBody.Balloon = &mem
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request body: %w", err)
	}

	endpoint := fmt.Sprintf("/nodes/%s/qemu", api.nodeID)
	baseRes, err := api.base.NewRequest(endpoint, "POST", bytes.NewReader(bodyBytes))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute the request
	_, err = baseRes.Execute()
	if err != nil {
		return 0, fmt.Errorf("request execution failed: %w", err)
	}

	return vmid, nil
}

func (api ProxmoxQEMUAPI) GetAllServers() (servers []interface_qemu.VMData, err error) {
	endpoint := fmt.Sprintf("/nodes/%s/qemu", api.nodeID)
	baseRes, err := api.base.NewRequest(endpoint, "GET", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resBody, err := baseRes.Execute()
	if err != nil {
		return nil, fmt.Errorf("request execution failed: %w", err)
	}

	var data []interface_qemu.VMData
	err = json.Unmarshal(resBody, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return data, nil
}

func (api ProxmoxQEMUAPI) GetServerByID(vmID int64) (server *interface_qemu.VMData, err error) {
	servers, err := api.GetAllServers()
	if err != nil {
		return
	}

	for _, serv := range servers {
		if serv.VMID == vmID {
			server = &serv
			return
		}
	}

	err = fmt.Errorf("server with ID %d not found", vmID)
	return
}

func (api ProxmoxQEMUAPI) GetServerByName(name string) (server *interface_qemu.VMData, err error) {
	servers, err := api.GetAllServers()
	if err != nil {
		return
	}

	for _, serv := range servers {
		if serv.Name == name {
			server = &serv
			return
		}
	}

	err = fmt.Errorf("server with name %s not found", name)
	return
}

func (api ProxmoxQEMUAPI) Delete(id int64) (err error) {
	endpoint := fmt.Sprintf("/nodes/%s/qemu/%d", api.nodeID, id)
	baseRes, err := api.base.NewRequest(endpoint, "DELETE", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = baseRes.Execute()
	if err != nil {
		return fmt.Errorf("request execution failed: %w", err)
	}

	return nil
}

func idExists(id int64, existingVMIDs []int64) bool {
	for _, existingID := range existingVMIDs {
		if existingID == id {
			return true
		}
	}
	return false
}

func generateNextAvailableVMID(startID int64, existingVMIDs []int64) int64 {
	for id := startID; id <= 199; id++ {
		if !idExists(id, existingVMIDs) {
			return id
		}
	}
	return -1
}
