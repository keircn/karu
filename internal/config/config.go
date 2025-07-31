package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

type Config struct {
	Player         string `json:"player"`
	PlayerArgs     string `json:"player_args"`
	Quality        string `json:"quality"`
	DownloadDir    string `json:"download_dir"`
	AutoPlayNext   bool   `json:"auto_play_next"`
	ShowSubtitles  bool   `json:"show_subtitles"`
	CacheTTL       int    `json:"cache_ttl_minutes"`
	RequestTimeout int    `json:"request_timeout_seconds"`
}

var DefaultConfig = Config{
	Player:         getDefaultPlayer(),
	PlayerArgs:     "",
	Quality:        "1080p",
	DownloadDir:    getDefaultDownloadDir(),
	AutoPlayNext:   false,
	ShowSubtitles:  true,
	CacheTTL:       15,
	RequestTimeout: 10,
}

func getDefaultPlayer() string {
	switch runtime.GOOS {
	case "darwin":
		return "iina"
	case "linux":
		return "mpv"
	case "windows":
		return "mpv.exe"
	default:
		return "mpv"
	}
}

func getDefaultDownloadDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./downloads"
	}
	return filepath.Join(homeDir, "Downloads", "karu")
}

func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	karuConfigDir := filepath.Join(configDir, "karu")
	if err := os.MkdirAll(karuConfigDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(karuConfigDir, "config.json"), nil
}

func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return &DefaultConfig, nil
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultConfig
		if err := Save(&config); err != nil {
			return &DefaultConfig, nil
		}
		return &config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return &DefaultConfig, err
	}

	config := DefaultConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return &DefaultConfig, err
	}

	if config.CacheTTL <= 0 {
		config.CacheTTL = DefaultConfig.CacheTTL
	}
	if config.RequestTimeout <= 0 {
		config.RequestTimeout = DefaultConfig.RequestTimeout
	}

	return &config, nil
}

func Save(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func (c *Config) Set(key, value string) error {
	switch key {
	case "player":
		c.Player = value
	case "player_args":
		c.PlayerArgs = value
	case "quality":
		c.Quality = value
	case "download_dir":
		c.DownloadDir = value
	case "auto_play_next":
		c.AutoPlayNext = value == "true"
	case "show_subtitles":
		c.ShowSubtitles = value == "true"
	case "cache_ttl_minutes":
		if ttl, err := strconv.Atoi(value); err == nil && ttl > 0 {
			c.CacheTTL = ttl
		}
	case "request_timeout_seconds":
		if timeout, err := strconv.Atoi(value); err == nil && timeout > 0 {
			c.RequestTimeout = timeout
		}
	}
	return Save(c)
}

func (c *Config) Get(key string) string {
	switch key {
	case "player":
		return c.Player
	case "player_args":
		return c.PlayerArgs
	case "quality":
		return c.Quality
	case "download_dir":
		return c.DownloadDir
	case "auto_play_next":
		if c.AutoPlayNext {
			return "true"
		}
		return "false"
	case "show_subtitles":
		if c.ShowSubtitles {
			return "true"
		}
		return "false"
	case "cache_ttl_minutes":
		return strconv.Itoa(c.CacheTTL)
	case "request_timeout_seconds":
		return strconv.Itoa(c.RequestTimeout)
	default:
		return ""
	}
}
