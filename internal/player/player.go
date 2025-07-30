package player

import (
	"os"
	"os/exec"
	"strings"

	"github.com/keircn/karu/internal/config"
)

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
