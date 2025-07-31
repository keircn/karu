package player

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/keircn/karu/internal/config"
)

type PlaybackInfo struct {
	ShowID   string
	Episodes []string
	Current  string
	VideoURL string
}

func Play(videoURL string) error {
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.DefaultConfig
	}

	args := []string{videoURL}
	if cfg.PlayerArgs != "" {
		playerArgs := strings.Fields(cfg.PlayerArgs)
		args = append(playerArgs, args...)
	}

	cmd := exec.Command(cfg.Player, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func PlayWithAutoNext(info *PlaybackInfo, getVideoURLFunc func(showID, episode string) (string, error)) error {
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.DefaultConfig
	}

	currentIndex := findEpisodeIndex(info.Episodes, info.Current)
	if currentIndex == -1 {
		return fmt.Errorf("current episode not found in episode list")
	}

	for i := currentIndex; i < len(info.Episodes); i++ {
		episode := info.Episodes[i]

		var videoURL string
		if i == currentIndex {
			videoURL = info.VideoURL
		} else {
			var err error
			videoURL, err = getVideoURLFunc(info.ShowID, episode)
			if err != nil {
				fmt.Printf("Error getting video URL for episode %s: %v\n", episode, err)
				break
			}
		}

		fmt.Printf("Playing episode: %s\n", episode)
		if cfg.AutoPlayNext && i > currentIndex {
			fmt.Printf("Auto-playing next episode...\n")
		}

		args := []string{videoURL}
		if cfg.PlayerArgs != "" {
			playerArgs := strings.Fields(cfg.PlayerArgs)
			args = append(playerArgs, args...)
		}

		cmd := exec.Command(cfg.Player, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return err
		}

		if !cfg.AutoPlayNext || i >= len(info.Episodes)-1 {
			break
		}
	}

	return nil
}

func findEpisodeIndex(episodes []string, target string) int {
	for i, episode := range episodes {
		if episode == target {
			return i
		}
		if episodeNum, err := strconv.Atoi(target); err == nil {
			if epNum, err := strconv.Atoi(episode); err == nil && epNum == episodeNum {
				return i
			}
		}
	}
	return -1
}
