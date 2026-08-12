package cmd

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gofrs/flock"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"raznar.id/proxmox-traffic-monitor/internal/monitor"
	"raznar.id/proxmox-traffic-monitor/internal/storage"
	"raznar.id/proxmox-traffic-monitor/internal/syncqueue"
	"raznar.id/proxmox-traffic-monitor/pkg/api/proxmox"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the traffic monitor daemon",
	Long:  `Start collecting and storing network traffic data for all VMs in the Proxmox cluster.`,
	Run:   start,
}

func start(cmd *cobra.Command, _ []string) {
	isVerbose, _ := cmd.Flags().GetBool("verbose")
	isDebug, _ := cmd.Flags().GetBool("debug")

	// Set global logging level
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	if isDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else if isVerbose {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Info().Msg("Starting up...")

	appConfig := loadConfig()
	lockPath := filepath.Join(filepath.Dir(appConfig.Filepath), "app.pid")

	// Ensure the directory exists
	appDir := filepath.Dir(lockPath)
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		log.Fatal().Err(err).Msgf("Failed to create lock directory: %s", appDir)
	}

	// Create file lock
	fileLock := flock.New(lockPath)
	locked, err := fileLock.TryLock()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to acquire file lock")
	}
	if !locked {
		log.Fatal().Msg("Another instance is already running")
	}

	defer fileLock.Unlock()

	syncqueue.GlobalJobQueue = syncqueue.New(appConfig.App.SyncQueueWorkers, appConfig.App.SyncQueueBuffer)

	// Initialize Proxmox API
	proxmoxAPI := proxmox.New()

	// Initialize Storage
	db, err := storage.New(appConfig.App.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	// Initialize Monitor
	fetchInterval := time.Duration(appConfig.App.FetchInterval) * time.Second
	m := monitor.New(&proxmoxAPI, db, fetchInterval, appConfig.App.RetentionPeriod)

	// Start monitor in a new goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.Start(ctx)

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down...")
}
