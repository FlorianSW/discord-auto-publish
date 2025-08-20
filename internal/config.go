package internal

import (
	"encoding/json"
	"log/slog"
	"os"
	"slices"
)

type Discord struct {
	Token string `json:"token"`
}

type Channels []string

func (c Channels) Contains(cid string) bool {
	if len(c) == 0 {
		return true
	}
	return slices.Contains(c, cid)
}

type Config struct {
	Discord  *Discord `json:"discord"`
	Channels Channels `json:"channels"`

	path string
}

func (c *Config) Save() error {
	config, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, config, 0655)
}

func NewConfig(path string, logger *slog.Logger) (*Config, error) {
	config, err := readConfig(path, logger)
	if err != nil {
		return config, err
	}

	return config, config.Save()
}

func readConfig(path string, logger *slog.Logger) (*Config, error) {
	var config Config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.Info("create-config")
		config = Config{}
	} else {
		logger.Info("read-existing-config")
		c, err := os.ReadFile(path)
		if err != nil {
			return &Config{}, err
		}
		err = json.Unmarshal(c, &config)
		if err != nil {
			return &Config{}, err
		}
	}
	config.path = path
	return &config, nil
}
