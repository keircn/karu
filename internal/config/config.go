package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/keircn/karu/pkg/errors"
	"github.com/keircn/karu/pkg/validation"
)

type Config struct {
	Player            string `json:"player"`
	PlayerArgs        string `json:"player_args"`
	Quality           string `json:"quality"`
	DownloadDir       string `json:"download_dir"`
	AutoPlayNext      bool   `json:"auto_play_next"`
	ShowSubtitles     bool   `json:"show_subtitles"`
	CacheTTL          int    `json:"cache_ttl_minutes"`
	RequestTimeout    int    `json:"request_timeout_seconds"`
	ConcurrentWorkers int    `json:"concurrent_workers"`
	PreloadEpisodes   int    `json:"preload_episodes"`
}

var DefaultConfig = Config{
	Player:            getDefaultPlayer(),
	PlayerArgs:        "",
	Quality:           "1080p",
	DownloadDir:       getDefaultDownloadDir(),
	AutoPlayNext:      false,
	ShowSubtitles:     true,
	CacheTTL:          15,
	RequestTimeout:    10,
	ConcurrentWorkers: 4,
	PreloadEpisodes:   5,
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
		return "", errors.Wrap(err, errors.ConfigError, "failed to get user config directory")
	}

	karuConfigDir := filepath.Join(configDir, "karu")
	if err := validation.EnsureDirectoryExists(karuConfigDir); err != nil {
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
		return &DefaultConfig, errors.Wrap(err, errors.ConfigError, "failed to read config file")
	}

	config := DefaultConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return &DefaultConfig, errors.Wrap(err, errors.ConfigError, "failed to parse config file")
	}

	if err := config.validate(); err != nil {
		return &DefaultConfig, err
	}

	config.applyDefaults()
	return &config, nil
}

func (c *Config) validate() error {
	if err := validation.ValidateNonEmptyString(c.Player, "player"); err != nil {
		return err
	}

	if err := validation.ValidateNonEmptyString(c.Quality, "quality"); err != nil {
		return err
	}

	if err := validation.ValidateNonEmptyString(c.DownloadDir, "download_dir"); err != nil {
		return err
	}

	if c.CacheTTL <= 0 {
		return errors.New(errors.ValidationError, "cache_ttl_minutes must be positive")
	}

	if c.RequestTimeout <= 0 {
		return errors.New(errors.ValidationError, "request_timeout_seconds must be positive")
	}

	if c.ConcurrentWorkers <= 0 {
		return errors.New(errors.ValidationError, "concurrent_workers must be positive")
	}

	if c.PreloadEpisodes < 0 {
		return errors.New(errors.ValidationError, "preload_episodes must be non-negative")
	}

	return nil
}

func (c *Config) applyDefaults() {
	if c.CacheTTL <= 0 {
		c.CacheTTL = DefaultConfig.CacheTTL
	}
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = DefaultConfig.RequestTimeout
	}
	if c.ConcurrentWorkers <= 0 {
		c.ConcurrentWorkers = DefaultConfig.ConcurrentWorkers
	}
	if c.PreloadEpisodes < 0 {
		c.PreloadEpisodes = DefaultConfig.PreloadEpisodes
	}
}

func Save(config *Config) error {
	if err := config.validate(); err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.Wrap(err, errors.ConfigError, "failed to marshal config")
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return errors.Wrap(err, errors.ConfigError, "failed to write config file")
	}

	return nil
}

func (c *Config) Set(key, value string) error {
	switch key {
	case "player":
		if err := validation.ValidateNonEmptyString(value, "player"); err != nil {
			return err
		}
		c.Player = value

	case "player_args":
		c.PlayerArgs = value

	case "quality":
		if err := validation.ValidateNonEmptyString(value, "quality"); err != nil {
			return err
		}
		c.Quality = value

	case "download_dir":
		if err := validation.ValidateNonEmptyString(value, "download_dir"); err != nil {
			return err
		}
		c.DownloadDir = value

	case "auto_play_next":
		c.AutoPlayNext = value == "true"

	case "show_subtitles":
		c.ShowSubtitles = value == "true"

	case "cache_ttl_minutes":
		ttl, err := validation.ValidatePositiveInt(value, "cache_ttl_minutes")
		if err != nil {
			return err
		}
		c.CacheTTL = ttl

	case "request_timeout_seconds":
		timeout, err := validation.ValidatePositiveInt(value, "request_timeout_seconds")
		if err != nil {
			return err
		}
		c.RequestTimeout = timeout

	case "concurrent_workers":
		workers, err := validation.ValidatePositiveInt(value, "concurrent_workers")
		if err != nil {
			return err
		}
		c.ConcurrentWorkers = workers

	case "preload_episodes":
		episodes, err := strconv.Atoi(value)
		if err != nil {
			return errors.Wrapf(err, errors.ValidationError, "invalid preload_episodes: must be a number")
		}
		if episodes < 0 {
			return errors.New(errors.ValidationError, "preload_episodes must be non-negative")
		}
		c.PreloadEpisodes = episodes

	default:
		return errors.New(errors.ValidationError, "unknown config key: "+key)
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
	case "concurrent_workers":
		return strconv.Itoa(c.ConcurrentWorkers)
	case "preload_episodes":
		return strconv.Itoa(c.PreloadEpisodes)
	default:
		return ""
	}
}
