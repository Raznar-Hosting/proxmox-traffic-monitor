package configs

import (
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Filepath string    `yaml:"-"`
	App      AppConfig `yaml:"app"`
}

var validate = validator.New()

// New loads the config from filePath, applies defaults, and validates it
func New(filePath string) (*Config, error) {
	config := &Config{
		Filepath: filePath,
	}

	// Load from file if exists
	if err := config.Reload(); err != nil {
		return nil, err
	}

	// Apply default values for missing fields
	config.applyDefaults()

	// Validate
	if err := validate.Struct(config); err != nil {
		return nil, err
	}

	return config, nil
}

// Reload reads the YAML config from file if it exists
func (c *Config) Reload() error {
	file, err := os.Open(c.Filepath)
	if os.IsNotExist(err) {
		// File does not exist, just use defaults
		c.applyDefaults()
		return c.Save()
	} else if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	return decoder.Decode(c)
}

// Save writes the current config to disk
func (c *Config) Save() error {
	file, err := os.Create(c.Filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	encoder.SetIndent(2) // Human-readable indentation

	return encoder.Encode(c)
}

// applyDefaults sets default values if they are not set
func (c *Config) applyDefaults() {
	if c.App.Database == "" {
		c.App.Database = "/var/lib/proxmox-traffic-monitor/data.db"
	}
	if c.App.RetentionPeriod == 0 {
		c.App.RetentionPeriod = 30
	}
	if c.App.FetchInterval < 1 {
		c.App.FetchInterval = 30
	}
	if c.App.SyncQueueWorkers == 0 {
		c.App.SyncQueueWorkers = 4
	}
	if c.App.SyncQueueBuffer == 0 {
		c.App.SyncQueueBuffer = 100
	}
}
