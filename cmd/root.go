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
	"raznar.id/proxmox-traffic-monitor/configs"
	"raznar.id/proxmox-traffic-monitor/internal/monitor"
	"raznar.id/proxmox-traffic-monitor/internal/storage"
	"raznar.id/proxmox-traffic-monitor/internal/syncqueue"
	"raznar.id/proxmox-traffic-monitor/pkg/api/proxmox"
)

var configPath string
var rootCmd = &cobra.Command{
	Use:   "proxmox-traffic-monitor",
	Short: "A tool to monitor Proxmox VM network traffic.",
	Long:  `A CLI tool to monitor network traffic for each VM in a Proxmox cluster, storing data in a local SQLite database.`,
	Run:   start,
}

func loadConfig() *configs.Config {
    cfg, err := configs.New(configPath)
    if err != nil {
        log.Fatal().Err(err).Msgf("Failed to load configuration from %s", configPath)
    }
    return cfg
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

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to acquire lock")
	}

	if !locked {
		log.Fatal().Msg("Another instance of the application is already running.")
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
	fetchInterval := time.Duration(appConfig.App.FetchInterval) * time.Minute
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

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}

func init() {
	rootCmd.Flags().BoolP("verbose", "v", false, "Enable info logging")
	rootCmd.Flags().BoolP("debug", "d", false, "Enable debug logging")
    rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "/var/lib/proxmox-traffic-monitor/config.yml", "Path to configuration file")

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
