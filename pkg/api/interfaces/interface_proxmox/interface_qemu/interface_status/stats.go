package interface_status

// Main data structure
type Data struct {
	Mem            int64                 `json:"mem"`
	Disk           int64                 `json:"disk"`
	CPUS           int                   `json:"cpus"`
	RunningQemu    string                `json:"running-qemu"`
	VMID           int                   `json:"vmid"`
	MaxMem         int64                 `json:"maxmem"`
	NetIn          int64                 `json:"netin"`
	Uptime         int64                 `json:"uptime"`
	CPU            float64               `json:"cpu"`
	DiskWrite      int64                 `json:"diskwrite"`
	QMPStatus      string                `json:"qmpstatus"`
	Nics           map[string]NICStats   `json:"nics"`
	MaxDisk        int64                 `json:"maxdisk"`
	DiskRead       int64                 `json:"diskread"`
	HA             HAStatus              `json:"ha"`
	ProxmoxSupport ProxmoxSupport        `json:"proxmox-support"`
	Name           string                `json:"name"`
	RunningMachine string                `json:"running-machine"`
	BlockStat      map[string]BlockStats `json:"blockstat"`
	PID            int                   `json:"pid"`
	Agent          int                   `json:"agent"`
	NetOut         int64                 `json:"netout"`
	Status         string                `json:"status"`
}

// NIC statistics structure
type NICStats struct {
	NetOut int64 `json:"netout"`
	NetIn  int64 `json:"netin"`
}

// HA status structure
type HAStatus struct {
	Managed int `json:"managed"`
}

// Proxmox support structure
type ProxmoxSupport struct {
	PBSDirtyBitmapSaveVM    bool   `json:"pbs-dirty-bitmap-savevm"`
	PBSDirtyBitmap          bool   `json:"pbs-dirty-bitmap"`
	BackupFleecing          bool   `json:"backup-fleecing"`
	PBSLibraryVersion       string `json:"pbs-library-version"`
	PBSDirtyBitmapMigration bool   `json:"pbs-dirty-bitmap-migration"`
	PBSMasterKey            bool   `json:"pbs-masterkey"`
	BackupMaxWorkers        bool   `json:"backup-max-workers"`
	QueryBitmapInfo         bool   `json:"query-bitmap-info"`
}

// Block statistics structure
type BlockStats struct {
	RdMerged                    int64   `json:"rd_merged"`
	InvalidFlushOperations      int64   `json:"invalid_flush_operations"`
	AccountInvalid              bool    `json:"account_invalid"`
	WrMerged                    int64   `json:"wr_merged"`
	FailedZoneAppendOperations  int64   `json:"failed_zone_append_operations"`
	InvalidUnmapOperations      int64   `json:"invalid_unmap_operations"`
	FailedFlushOperations       int64   `json:"failed_flush_operations"`
	TimedStats                  []int64 `json:"timed_stats"` // Assuming this is an array of integers
	ZoneAppendOperations        int64   `json:"zone_append_operations"`
	IdleTimeNS                  int64   `json:"idle_time_ns"`
	UnmapMerged                 int64   `json:"unmap_merged"`
	FailedRdOperations          int64   `json:"failed_rd_operations"`
	WrHighestOffset             int64   `json:"wr_highest_offset"`
	InvalidZoneAppendOperations int64   `json:"invalid_zone_append_operations"`
	FailedUnmapOperations       int64   `json:"failed_unmap_operations"`
	AccountFailed               bool    `json:"account_failed"`
	UnmapTotalTimeNS            int64   `json:"unmap_total_time_ns"`
	WrTotalTimeNS               int64   `json:"wr_total_time_ns"`
	RdTotalTimeNS               int64   `json:"rd_total_time_ns"`
	FlushTotalTimeNS            int64   `json:"flush_total_time_ns"`
	RdBytes                     int64   `json:"rd_bytes"`
	WrBytes                     int64   `json:"wr_bytes"`
	InvalidRdOperations         int64   `json:"invalid_rd_operations"`
	ZoneAppendBytes             int64   `json:"zone_append_bytes"`
	ZoneAppendMerged            int64   `json:"zone_append_merged"`
	FailedWrOperations          int64   `json:"failed_wr_operations"`
	ZoneAppendTotalTimeNS       int64   `json:"zone_append_total_time_ns"`
	FlushOperations             int64   `json:"flush_operations"`
	UnmapOperations             int64   `json:"unmap_operations"`
	UnmapBytes                  int64   `json:"unmap_bytes"`
	InvalidWrOperations         int64   `json:"invalid_wr_operations"`
	RdOperations                int64   `json:"rd_operations"`
	WrOperations                int64   `json:"wr_operations"`
}
