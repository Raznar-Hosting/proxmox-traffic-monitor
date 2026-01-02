package monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"raznar.id/proxmox-traffic-monitor/internal/storage"
	"raznar.id/proxmox-traffic-monitor/internal/syncqueue"
	"raznar.id/proxmox-traffic-monitor/pkg/api/proxmox"
)

type Monitor struct {
	api         *proxmox.ProxmoxAPI
	storage     *storage.Storage
	interval    time.Duration
	retention   int
	prevTraffic map[string]struct {
		NetIn  uint64
		NetOut uint64
	}
}

func New(api *proxmox.ProxmoxAPI, s *storage.Storage, interval time.Duration, retention int) *Monitor {
	return &Monitor{
		api:       api,
		storage:   s,
		interval:  interval,
		retention: retention,
		prevTraffic: make(map[string]struct {
			NetIn  uint64
			NetOut uint64
		}),
	}
}
func (m *Monitor) Start(ctx context.Context) {
	m.initTraffic()

	log.Info().Msg("Starting traffic monitor...")
	ticker := time.NewTicker(m.interval)
	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	defer cleanupTicker.Stop()

	// Initial run
	log.Debug().Msg("Enqueueing initial traffic collection job")
	m.enqueueJob()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Traffic monitor stopped via context")
			return
		case <-ticker.C:
			log.Debug().Msg("Enqueueing scheduled traffic collection job")
			m.enqueueJob()
		case <-cleanupTicker.C:
			log.Debug().Msg("Enqueueing daily cleanup job")
			m.cleanup()
		}
	}
}

func (m *Monitor) enqueueJob() {
	log.Debug().Msg("Adding job to sync queue")
	syncqueue.GlobalJobQueue.Add(m.collectTraffic, 3, 5*time.Second, 50*time.Second)
}
func (m *Monitor) collectTraffic(ctx context.Context) error {
	nodes, err := m.api.NodesAPI().GetNodes()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get Proxmox nodes")
		return fmt.Errorf("failed to get nodes: %w", err)
	}

	now := time.Now()

	for _, node := range nodes {
		nodeID := node.GetId()

		qemuAPI := m.api.QEMUAPI(nodeID)
		vms, err := qemuAPI.GetAllServers()
		if err != nil {
			log.Warn().Err(err).Str("node", nodeID).Msg("Failed to get VMs for node")
			continue
		}

		for _, vm := range vms {
			stats, err := qemuAPI.StatusAPI().Stats(vm.VMID)
			if err != nil {
				log.Warn().Err(err).
					Int64("vmid", vm.VMID).
					Str("vm_name", vm.Name).
					Msg("Failed to get stats for VM")
				continue
			}

			id := fmt.Sprintf("%s-%d-%s", nodeID, vm.VMID, vm.Name)

			currentNetIn := uint64(stats.NetIn)
			currentNetOut := uint64(stats.NetOut)

			var deltaNetIn, deltaNetOut uint64

			if prev, ok := m.prevTraffic[id]; ok {
				if currentNetIn >= prev.NetIn {
					deltaNetIn = currentNetIn - prev.NetIn
				} else {
					deltaNetIn = currentNetIn
					log.Warn().Str("vm", id).
						Uint64("prev_in", prev.NetIn).
						Uint64("current_in", currentNetIn).
						Msg("NetIn counter reset detected")
				}

				if currentNetOut >= prev.NetOut {
					deltaNetOut = currentNetOut - prev.NetOut
				} else {
					deltaNetOut = currentNetOut
					log.Warn().Str("vm", id).
						Uint64("prev_out", prev.NetOut).
						Uint64("current_out", currentNetOut).
						Msg("NetOut counter reset detected")
				}
			} else {
				// First run after start/restart
				deltaNetIn = 0
				deltaNetOut = 0
				log.Debug().Str("vm", id).
					Msg("First traffic collection, skipping delta")
			}

			// Update cache
			m.prevTraffic[id] = struct {
				NetIn  uint64
				NetOut uint64
			}{
				NetIn:  currentNetIn,
				NetOut: currentNetOut,
			}

			// Skip zero delta
			if deltaNetIn == 0 && deltaNetOut == 0 {
				continue
			}

			record := storage.TrafficRecord{
				ID:        id,
				NodeID:    nodeID,
				VMID:      vm.VMID,
				In:        deltaNetIn,
				Out:       deltaNetOut,
				Timestamp: &now,
			}

			if err := m.storage.UpdateTraffic(record); err != nil {
				log.Warn().Err(err).Str("id", id).
					Msg("Failed to update traffic in storage")
			}
		}
	}

	return nil
}
func (m *Monitor) initTraffic() {
	log.Info().Msg("Running initial traffic...")

	nodes, err := m.api.NodesAPI().GetNodes()
	if err != nil {
		log.Warn().Err(err).
			Msg("Failed to get Proxmox nodes for initial kickstarter")
		return
	}

	for _, node := range nodes {
		nodeID := node.GetId()
		qemuAPI := m.api.QEMUAPI(nodeID)

		vms, err := qemuAPI.GetAllServers()
		if err != nil {
			log.Warn().Err(err).
				Str("node", nodeID).
				Msg("Failed to get VMs for kickstarter")
			continue
		}

		for _, vm := range vms {
			stats, err := qemuAPI.StatusAPI().Stats(vm.VMID)
			if err != nil {
				log.Warn().Err(err).
					Int64("vmid", vm.VMID).
					Str("vm_name", vm.Name).
					Msg("Failed to get stats for kickstarter")
				continue
			}

			currentNetIn := uint64(stats.NetIn)
			currentNetOut := uint64(stats.NetOut)

			if currentNetIn == 0 && currentNetOut == 0 {
				continue
			}

			id := fmt.Sprintf("%s-%d-%s", nodeID, vm.VMID, vm.Name)

			// Check if DB has any record for this VM
			hasRecord, err := m.storage.ExistsTraffic(id)
			if err != nil {
				log.Warn().Err(err).
					Str("vm", id).
					Msg("Failed to check DB for kickstarter")
				continue
			}

			if !hasRecord {
				now := time.Now()

				// Initialize baseline record
				record := storage.TrafficRecord{
					ID:        id,
					NodeID:    nodeID,
					VMID:      vm.VMID,
					In:        currentNetIn,
					Out:       currentNetOut,
					Timestamp: &now,
				}

				if err := m.storage.UpdateTraffic(record); err != nil {
					log.Warn().Err(err).
						Str("id", id).
						Msg("Failed to update traffic in storage")
				}

				// Populate in-memory cache
				m.prevTraffic[id] = struct {
					NetIn  uint64
					NetOut uint64
				}{
					NetIn:  currentNetIn,
					NetOut: currentNetOut,
				}

				log.Debug().
					Str("vm", id).
					Uint64("net_in", currentNetIn).
					Uint64("net_out", currentNetOut).
					Msg("Kickstarter: initialized baseline traffic")
			}
		}
	}
}

func (m *Monitor) cleanup() {
	log.Info().Msg("Running retention cleanup...")
	deleted, err := m.storage.CleanupOldRecords(m.retention)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to cleanup old records")
	} else {
		log.Debug().Int64("deleted_records", deleted).Msg("Retention cleanup completed")
	}
}
