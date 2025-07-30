package scraper

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ProgressWriter struct {
	total      int64
	downloaded int64
	filename   string
	startTime  time.Time
}

func NewProgressWriter(total int64, filename string) *ProgressWriter {
	return &ProgressWriter{
		total:     total,
		filename:  filename,
		startTime: time.Now(),
	}
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.downloaded += int64(n)

	if pw.total > 0 {
		percent := float64(pw.downloaded) / float64(pw.total) * 100
		elapsed := time.Since(pw.startTime)

		if pw.downloaded > 0 {
			speed := float64(pw.downloaded) / elapsed.Seconds()
			remaining := time.Duration(float64(pw.total-pw.downloaded)/speed) * time.Second

			fmt.Printf("\r%s: %.1f%% (%.2f MB/%.2f MB) [%.2f MB/s] ETA: %v",
				pw.filename,
				percent,
				float64(pw.downloaded)/(1024*1024),
				float64(pw.total)/(1024*1024),
				speed/(1024*1024),
				remaining.Round(time.Second))
		}
	} else {
		fmt.Printf("\r%s: %.2f MB downloaded",
			pw.filename,
			float64(pw.downloaded)/(1024*1024))
	}

	return n, nil
}

func DownloadEpisodeWithProgress(showID, episode, outputPath string) error {
	videoURL, err := GetVideoURL(showID, episode)
	if err != nil {
		return fmt.Errorf("failed to get video URL: %w", err)
	}

	if videoURL == "" {
		return fmt.Errorf("no video URL found for episode %s", episode)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}

	resp, err := http.Get(videoURL)
	if err != nil {
		return fmt.Errorf("failed to download video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download video: status %d", resp.StatusCode)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	contentLength := resp.ContentLength
	filename := filepath.Base(outputPath)

	fmt.Printf("Starting download: %s\n", filename)

	var writer io.Writer = out
	if contentLength > 0 {
		pw := NewProgressWriter(contentLength, filename)
		writer = io.MultiWriter(out, pw)
	}

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write video data: %w", err)
	}

	fmt.Printf("\nDownload completed: %s\n", outputPath)
	return nil
}
