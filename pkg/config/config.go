package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	configPath = "./config.yaml"
)

type Config struct {
	Env    string `yaml:"env" env-default:"development"`
	Server struct {
		Enabled bool   `yaml:"enabled" env-default:"true"`
		Port    string `yaml:"port" env-default:"8080"`
		APIKey  string `yaml:"api_key"`
	} `yaml:"server"`
	Database struct {
		URL             string `yaml:"url"`
		MaxOpenConns    int    `yaml:"max_open_conns" env-default:"25"`
		MaxIdleConns    int    `yaml:"max_idle_conns" env-default:"25"`
		ConnMaxLifetime int    `yaml:"conn_max_lifetime" env-default:"5"`
	} `yaml:"database"`
	AI struct {
		Enabled     bool    `yaml:"enabled" env-default:"true"`
		Provider    string  `yaml:"provider" env-default:"openai"`
		Model       string  `yaml:"model" env-default:"gpt-4o-mini"`
		Temperature float64 `yaml:"temperature" env-default:"0.3"`
		APIKey      string  `yaml:"api_key"`
	} `yaml:"ai"`
	Queue struct {
		WorkerCount    int    `yaml:"worker_count" env-default:"1"`
		PollIntervalMs int    `yaml:"poll_interval_ms" env-default:"1000"`
		Type           string `yaml:"type" env-default:"inmemory"`
		BufferSize     int    `yaml:"buffer_size" env-default:"100"`
	} `yaml:"queue"`
}

func Load(path ...string) (*Config, error) {
	cfg := Config{}

	if len(path) > 0 {
		configPath = path[0]
	}

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		return &cfg, err
	}

	// Defaults
	if cfg.AI.APIKey == "" {
		cfg.AI.APIKey = os.Getenv("OPENAI_API_KEY")
	}

	return &cfg, nil
}
