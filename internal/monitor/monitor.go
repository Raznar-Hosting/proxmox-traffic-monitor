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
	log.Debug().Int("nodes_count", len(nodes)).Msg("Fetched nodes from Proxmox")

	now := time.Now()
	date := now.Format("02-01-2006")

	for _, node := range nodes {
		nodeID := node.GetId()
		log.Debug().Str("node", nodeID).Msg("Processing node")

		qemuAPI := m.api.QEMUAPI(nodeID)
		vms, err := qemuAPI.GetAllServers()
		if err != nil {
			log.Warn().Err(err).Str("node", nodeID).Msg("Failed to get VMs for node")
			continue
		}

		log.Debug().Int("vm_count", len(vms)).Str("node", nodeID).Msg("VMs fetched for node")

		for _, vm := range vms {
			stats, err := qemuAPI.StatusAPI().Stats(vm.VMID)
			if err != nil {
				log.Warn().Err(err).Int64("vmid", vm.VMID).Str("vm_name", vm.Name).Msg("Failed to get stats for VM")
				continue
			}

			id := fmt.Sprintf("%s-%d-%s", nodeID, vm.VMID, vm.Name)
			currentNetIn := uint64(stats.NetIn)
			currentNetOut := uint64(stats.NetOut)

			var deltaNetIn uint64
			var deltaNetOut uint64

			if prev, ok := m.prevTraffic[id]; ok {
				// Calculate delta
				if currentNetIn >= prev.NetIn {
					deltaNetIn = currentNetIn - prev.NetIn
				} else {
					// Counter reset or overflow, assume current value is the new delta from 0
					deltaNetIn = currentNetIn
					log.Warn().Str("vm", id).Uint64("prev_in", prev.NetIn).Uint64("current_in", currentNetIn).Msg("NetIn counter reset detected, assuming new base.")
				}

				if currentNetOut >= prev.NetOut {
					deltaNetOut = currentNetOut - prev.NetOut
				} else {
					// Counter reset or overflow, assume current value is the new delta from 0
					deltaNetOut = currentNetOut
					log.Warn().Str("vm", id).Uint64("prev_out", prev.NetOut).Uint64("current_out", currentNetOut).Msg("NetOut counter reset detected, assuming new base.")
				}
			} else {
				// First collection for this VM, or after monitor restart.
				// We don't have previous data to calculate delta.
				// For this interval, delta is 0, but we populate the cache for next interval.
				deltaNetIn = 0
				deltaNetOut = 0
				log.Debug().Str("vm", id).Msg("First traffic collection for VM, skipping delta calculation for this interval.")
			}

			// Update cache with current cumulative values for the next interval
			m.prevTraffic[id] = struct {
				NetIn  uint64
				NetOut uint64
			}{
				NetIn:  currentNetIn,
				NetOut: currentNetOut,
			}

			// Required: Always skip records with zero traffic
			if deltaNetIn == 0 && deltaNetOut == 0 {
				log.Debug().Str("vm", id).Msg("Skipping record due to zero traffic delta.")
				continue // Skip storing this record
			}

			record := storage.TrafficRecord{
				ID:        id,
				Date:      date,
				In:        deltaNetIn,
				Out:       deltaNetOut,
				NodeID:    nodeID,
				VMID:      vm.VMID,
				Timestamp: &now,
			}

			log.Debug().
				Str("vm", id).
				Uint64("net_in_delta", record.In).
				Uint64("net_out_delta", record.Out).
				Time("timestamp", now).
				Msg("Traffic stats collected with delta")

			if err := m.storage.UpdateTraffic(record); err != nil {
				log.Warn().Err(err).Str("id", id).Msg("Failed to update traffic in storage")
			} else {
				log.Debug().Str("id", id).Msg("Traffic successfully updated in storage")
			}
		}
	}

	return nil
}
func (m *Monitor) initTraffic() {
	log.Info().Msg("Running initial traffic...")

	nodes, err := m.api.NodesAPI().GetNodes()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get Proxmox nodes for initial kickstarter")
		return
	}

	for _, node := range nodes {
		nodeID := node.GetId()
		qemuAPI := m.api.QEMUAPI(nodeID)

		vms, err := qemuAPI.GetAllServers()
		if err != nil {
			log.Warn().Err(err).Str("node", nodeID).Msg("Failed to get VMs for kickstarter")
			continue
		}

		for _, vm := range vms {
			stats, err := qemuAPI.StatusAPI().Stats(vm.VMID)
			if err != nil {
				log.Warn().Err(err).Int64("vmid", vm.VMID).Str("vm_name", vm.Name).Msg("Failed to get stats for kickstarter")
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
				log.Warn().Err(err).Str("vm", id).Msg("Failed to check DB for kickstarter")
				continue
			}

			if !hasRecord {
				now := time.Now()
				date := now.Format("02-01-2006")
				// No previous record in DB -> initialize traffic record
				record := storage.TrafficRecord{
					ID:        id,
					Date:      date,
					In:        currentNetIn,
					Out:       currentNetOut,
					NodeID:    nodeID,
					VMID:      vm.VMID,
					Timestamp: &now,
				}

				if err := m.storage.UpdateTraffic(record); err != nil {
					log.Warn().Err(err).Str("id", id).Msg("Failed to update traffic in storage")
				} else {
					log.Debug().Str("id", id).Msg("Traffic successfully updated in storage")
				}

				// Populate in-memory prevTraffic cache for immediate delta calculation
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
					Msg("Kickstarter: initialized delta with current traffic and cache")
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
