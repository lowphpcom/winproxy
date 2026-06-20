package main

import (
	"encoding/json"
	"os"
)

const configFile = "winproxy.json"

type Config struct {
	ListenHost string
	ListenPort string
	TargetHost string
	TargetPort string
	SocksHost  string
	SocksPort  string
	Username   string
	Password   string
	Language   string
}

func defaultConfig() Config {
	return Config{
		ListenHost: "127.0.0.1",
		ListenPort: "757",
		Language:   "zh-CN",
	}
}

func loadConfig() Config {
	cfg := defaultConfig()
	data, err := os.ReadFile(configFile)
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, &cfg)
	if cfg.ListenHost == "" {
		cfg.ListenHost = "127.0.0.1"
	}
	if cfg.ListenPort == "" {
		cfg.ListenPort = "757"
	}
	if cfg.Language == "" {
		cfg.Language = "zh-CN"
	}
	return cfg
}

func saveConfig(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0644)
}
