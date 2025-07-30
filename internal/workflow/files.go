package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/keircn/karu/internal/config"
)

type DownloadInfo struct {
	Path     string
	Name     string
	Size     int64
	Modified time.Time
}

type FileManager struct {
	Config *config.Config
}

func NewFileManager() (*FileManager, error) {
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.DefaultConfig
	}
	return &FileManager{Config: cfg}, nil
}

func (fm *FileManager) ListDownloads() ([]DownloadInfo, error) {
	files, err := filepath.Glob(filepath.Join(fm.Config.DownloadDir, "*.mp4"))
	if err != nil {
		return nil, fmt.Errorf("listing downloads: %w", err)
	}

	var downloads []DownloadInfo
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		downloads = append(downloads, DownloadInfo{
			Path:     file,
			Name:     filepath.Base(file),
			Size:     info.Size(),
			Modified: info.ModTime(),
		})
	}

	return downloads, nil
}

func (fm *FileManager) CleanDownloads() (int, error) {
	files, err := filepath.Glob(filepath.Join(fm.Config.DownloadDir, "*.mp4"))
	if err != nil {
		return 0, fmt.Errorf("listing downloads: %w", err)
	}

	removed := 0
	for _, file := range files {
		if err := os.Remove(file); err == nil {
			removed++
		}
	}

	return removed, nil
}
