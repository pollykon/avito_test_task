package cmd

import (
	"github.com/caarlos0/env/v7"
)

type Config struct {
	Database         DatabaseConfig
	CronTimeInterval CronTimeIntervalConfig
	BatchSize        DeleteBatchSizeConfig
	CSV              CSVConfig
}

type DatabaseConfig struct {
	Host     string `env:"PG_HOST,required"`
	Port     string `env:"PG_PORT,required"`
	User     string `env:"PG_USER,required"`
	Password string `env:"PG_PASSWORD,required"`
	Name     string `env:"PG_DATABASE_NAME,required"`
}

type CSVConfig struct {
	LogCSVDirectory string `env:"LOGS_CSV_DIRECTORY,required"`
}

type CronTimeIntervalConfig struct {
	DeleteSegments    int `env:"TIME_INTERVAL_DELETE_SEGMENTS,required"`
	DeleteTTLSegments int `env:"TIME_INTERVAL_DELETE_TTL_SEGMENTS,required"`
	DeleteLogs        int `env:"TIME_INTERVAL_DELETE_LOGS,required"`
}

type DeleteBatchSizeConfig struct {
	Segments    int64 `env:"BATCH_SIZE_SEGMENTS,required"`
	TTLSegments int64 `env:"BATCH_SIZE_TTL_SEGMENTS,required"`
	Logs        int64 `env:"BATCH_SIZE_LOGS_MINUTES,required"`
}

func Load() (*Config, error) {
	cfg := Config{}

	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
