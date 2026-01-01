package interface_qemu

import "strings"

type VMData struct {
	PID        int64   `json:"pid"`
	Name       string  `json:"name"`
	Status     string  `json:"status"`
	NetOut     int64   `json:"netout"`
	MaxDisk    int64   `json:"maxdisk"`
	DiskRead   int64   `json:"diskread"`
	Uptime     int64   `json:"uptime"`
	NetIn      int64   `json:"netin"`
	DiskWrite  int64   `json:"diskwrite"`
	CPU        float64 `json:"cpu"`
	CPUs       int64   `json:"cpus"`
	Disk       int64   `json:"disk"`
	Mem        int64   `json:"mem"`
	VMID       int64   `json:"vmid"`
	MaxMem     int64   `json:"maxmem"`
	BalloonMin int64   `json:"balloon_min,omitempty"` // Optional field
	Shares     int64   `json:"shares,omitempty"`      // Optional field
}

type CreateVMConfig struct {
	VMConfig
	VMID int64 `json:"vmid"`
}

type EmptyArgsConfig struct {
	Args string `json:"args"`
}

type VMConfig struct {
	Args       *string `json:"args,omitempty"`
	Shares     int64  `json:"shares,omitempty"`
	OSType     string `json:"ostype,omitempty"`
	Digest     string `json:"digest,omitempty"`
	IDE2       string `json:"ide2,omitempty"`
	SCSIHW     string `json:"scsihw,omitempty"`
	Bios       string `json:"bios,omitempty"`
	IPConfig0  string `json:"ipconfig0,omitempty"`
	VGA        string `json:"vga,omitempty"`
	Serial0    string `json:"serial0,omitempty"`
	Scsi0      string `json:"scsi0,omitempty"`
	Scsi1      string `json:"scsi1,omitempty"`
	Scsi2      string `json:"scsi2,omitempty"`
	Scsi3      string `json:"scsi3,omitempty"`
	Scsi4      string `json:"scsi4,omitempty"`
	Scsi5      string `json:"scsi5,omitempty"`
	Scsi6      string `json:"scsi6,omitempty"`
	Scsi7      string `json:"scsi7,omitempty"`
	Scsi8      string `json:"scsi8,omitempty"`
	Scsi9      string `json:"scsi9,omitempty"`
	Scsi10     string `json:"scsi10,omitempty"`
	Scsi11     string `json:"scsi11,omitempty"`
	Scsi12     string `json:"scsi12,omitempty"`
	Scsi13     string `json:"scsi13,omitempty"`
	Scsi14     string `json:"scsi14,omitempty"`
	Scsi15     string `json:"scsi15,omitempty"`
	Scsi16     string `json:"scsi16,omitempty"`
	Scsi17     string `json:"scsi17,omitempty"`
	Scsi18     string `json:"scsi18,omitempty"`
	Scsi19     string `json:"scsi19,omitempty"`
	Numa       int    `json:"numa,omitempty"`
	Name       string `json:"name,omitempty"`
	CIUser     string `json:"ciuser,omitempty"`
	Net0       string `json:"net0,omitempty"`
	Net1       string `json:"net1,omitempty"`
	Net2       string `json:"net2,omitempty"`
	Net3       string `json:"net3,omitempty"`
	Net4       string `json:"net4,omitempty"`
	Net5       string `json:"net5,omitempty"`
	Net6       string `json:"net6,omitempty"`
	Net7       string `json:"net7,omitempty"`
	Net8       string `json:"net8,omitempty"`
	Net9       string `json:"net9,omitempty"`
	Net10      string `json:"net10,omitempty"`
	Net11      string `json:"net11,omitempty"`
	Net12      string `json:"net12,omitempty"`
	Net13      string `json:"net13,omitempty"`
	Net14      string `json:"net14,omitempty"`
	Net15      string `json:"net15,omitempty"`
	Net16      string `json:"net16,omitempty"`
	Net17      string `json:"net17,omitempty"`
	Net18      string `json:"net18,omitempty"`
	Net19      string `json:"net19,omitempty"`
	Agent      string `json:"agent,omitempty"`
	SSHKeys    string `json:"sshkeys,omitempty"`
	Meta       string `json:"meta,omitempty"`
	Boot       string `json:"boot,omitempty"`
	Cores      int64  `json:"cores,omitempty"`
	Nameserver string `json:"nameserver,omitempty"`
	Memory     string `json:"memory,omitempty"`
	Sockets    int    `json:"sockets,omitempty"`
	VMGenID    string `json:"vmgenid,omitempty"`
	Balloon    *int   `json:"balloon,omitempty"`
	CPU        string `json:"cpu,omitempty"`
	SMBIOS1    string `json:"smbios1,omitempty"`
	Unused0    string `json:"unused0,omitempty"`
	Unused1    string `json:"unused1,omitempty"`
	Unused2    string `json:"unused2,omitempty"`
	Unused3    string `json:"unused3,omitempty"`
	Unused4    string `json:"unused4,omitempty"`
	Unused5    string `json:"unused5,omitempty"`
	Unused6    string `json:"unused6,omitempty"`
	Unused7    string `json:"unused7,omitempty"`
	Unused8    string `json:"unused8,omitempty"`
	Unused9    string `json:"unused9,omitempty"`
	Unused10   string `json:"unused10,omitempty"`
	Unused11   string `json:"unused11,omitempty"`
	Unused12   string `json:"unused12,omitempty"`
	Unused13   string `json:"unused13,omitempty"`
	Unused14   string `json:"unused14,omitempty"`
	Unused15   string `json:"unused15,omitempty"`
	Unused16   string `json:"unused16,omitempty"`
	Unused17   string `json:"unused17,omitempty"`
	Unused18   string `json:"unused18,omitempty"`
	Unused19   string `json:"unused19,omitempty"`
}

// GetNetworks returns the list of network configurations from Net0 to Net19.
func (c *VMConfig) GetNetworks() []string {
	nets := []string{
		c.Net0, c.Net1, c.Net2, c.Net3, c.Net4,
		c.Net5, c.Net6, c.Net7, c.Net8, c.Net9,
		c.Net10, c.Net11, c.Net12, c.Net13, c.Net14,
		c.Net15, c.Net16, c.Net17, c.Net18, c.Net19,
	}
	// Filter out empty strings
	var result []string
	for _, net := range nets {
		if net != "" {
			result = append(result, net)
		}
	}
	return result
}

// GetUnusedDisks returns the list of unused disks from Unused0 to Unused19.
func (c *VMConfig) GetUnusedDisks() []string {
	unusedDisks := []string{
		c.Unused0, c.Unused1, c.Unused2, c.Unused3, c.Unused4,
		c.Unused5, c.Unused6, c.Unused7, c.Unused8, c.Unused9,
		c.Unused10, c.Unused11, c.Unused12, c.Unused13, c.Unused14,
		c.Unused15, c.Unused16, c.Unused17, c.Unused18, c.Unused19,
	}
	// Filter out empty strings
	var result []string
	for _, disk := range unusedDisks {
		if disk != "" {
			result = append(result, disk)
		}
	}
	return result
}

// GetUsedDisks returns the list of used disks from Scsi0 to Scsi19.
func (c *VMConfig) GetUsedDisks() []string {
	usedDisks := []string{
		c.Scsi0, c.Scsi1, c.Scsi2, c.Scsi3, c.Scsi4,
		c.Scsi5, c.Scsi6, c.Scsi7, c.Scsi8, c.Scsi9,
		c.Scsi10, c.Scsi11, c.Scsi12, c.Scsi13, c.Scsi14,
		c.Scsi15, c.Scsi16, c.Scsi17, c.Scsi18, c.Scsi19,
	}
	// Filter out empty strings
	var result []string
	for _, disk := range usedDisks {
		if disk != "" {
			result = append(result, disk)
		}
	}
	return result
}

// find the disk using contains
func (c *VMConfig) FindUsedDisk(value string) (disk string, index int) {
	// Define a list of used disks from scsi0 to scsi20
	scsiDisks := []string{
		c.Scsi0, c.Scsi1, c.Scsi2, c.Scsi3, c.Scsi4,
		c.Scsi5, c.Scsi6, c.Scsi7, c.Scsi8, c.Scsi9,
		c.Scsi10, c.Scsi11, c.Scsi12, c.Scsi13, c.Scsi14,
		c.Scsi15, c.Scsi16, c.Scsi17, c.Scsi18, c.Scsi19,
	}

	if !strings.HasSuffix(value, ".qcow2") {
		value = value + ".qcow2"
	}

	// Loop through each SCSI disk and check for the value
	for i, d := range scsiDisks {
		if strings.Contains(d, value) {
			return d, i // Return the disk and its index
		}
	}
	return "", -1 // Return an empty string and -1 if no disk found
}

// GetEmptyUsedSlotDisk returns the first empty slot for used disks.
func (c *VMConfig) GetEmptyUsedSlotDisk() int {
	scsiDisks := []string{
		c.Scsi0, c.Scsi1, c.Scsi2, c.Scsi3, c.Scsi4,
		c.Scsi5, c.Scsi6, c.Scsi7, c.Scsi8, c.Scsi9,
		c.Scsi10, c.Scsi11, c.Scsi12, c.Scsi13, c.Scsi14,
		c.Scsi15, c.Scsi16, c.Scsi17, c.Scsi18, c.Scsi19,
	}
	// Look for the first empty slot in the scsiDisks array
	for i, disk := range scsiDisks {
		if disk == "" {
			return i
		}
	}
	return -1 // Return -1 if no empty slot found
}

// FindUnusedDisk finds an unused disk.
func (c *VMConfig) FindUnusedDisk(value string) (disk string, index int) {
	unusedDisks := []string{
		c.Unused0, c.Unused1, c.Unused2, c.Unused3, c.Unused4,
		c.Unused5, c.Unused6, c.Unused7, c.Unused8, c.Unused9,
		c.Unused10, c.Unused11, c.Unused12, c.Unused13, c.Unused14,
		c.Unused15, c.Unused16, c.Unused17, c.Unused18, c.Unused19,
	}

	if !strings.HasSuffix(value, ".qcow2") {
		value = value + ".qcow2"
	}

	for i, disk := range unusedDisks {
		if strings.Contains(disk, value) {
			return disk, i
		}
	}
	return "", -1 // Return empty and -1 if no match
}

// SetNetworks sets the network configurations for the VM.
func (c *VMConfig) SetNetworks(nw []string) {
	for i, net := range nw {
		switch i {
		case 0:
			c.Net0 = net
		case 1:
			c.Net1 = net
		case 2:
			c.Net2 = net
		case 3:
			c.Net3 = net
		case 4:
			c.Net4 = net
		case 5:
			c.Net5 = net
		case 6:
			c.Net6 = net
		case 7:
			c.Net7 = net
		case 8:
			c.Net8 = net
		case 9:
			c.Net9 = net
		case 10:
			c.Net10 = net
		case 11:
			c.Net11 = net
		case 12:
			c.Net12 = net
		case 13:
			c.Net13 = net
		case 14:
			c.Net14 = net
		case 15:
			c.Net15 = net
		case 16:
			c.Net16 = net
		case 17:
			c.Net17 = net
		case 18:
			c.Net18 = net
		case 19:
			c.Net19 = net
		}
	}
}
