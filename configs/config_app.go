package configs



type AppConfig struct {
	// Job queue settings
	SyncQueueWorkers int `yaml:"sync_queue_workers"` // number of workers
	SyncQueueBuffer  int `yaml:"sync_queue_buffer"`  // optional buffer size

	// Traffic monitoring settings
	Database        string `yaml:"database"`         // path to sqlite database
	RetentionPeriod int    `yaml:"retention-period"` // in days
	FetchInterval   int    `yaml:"fetch-interval"`   // in minutes


}
