package cmd

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"raznar.id/proxmox-traffic-monitor/configs"
)

var configPath string
var jsonOutput bool

var rootCmd = &cobra.Command{
	Use:   "pmx-tm",
	Short: "A tool to monitor Proxmox VM network traffic.",
	Long:  `A CLI tool to monitor network traffic for each VM in a Proxmox cluster, storing data in a local SQLite database.`,
}

func loadConfig() *configs.Config {
	// Ensure the directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		log.Fatal().Err(err).Msgf("Failed to create config directory: %s", configDir)
	}

	cfg, err := configs.New(configPath)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to load configuration from %s", configPath)
	}
	return cfg
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}

func init() {
	startCmd.Flags().BoolP("verbose", "v", false, "Enable info logging")
	startCmd.Flags().BoolP("debug", "d", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "/var/lib/proxmox-traffic-monitor/config.yml", "Path to configuration file")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	rootCmd.AddCommand(startCmd)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
