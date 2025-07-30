package player

import (
	"os"
	"os/exec"
	"runtime"
)

func getPlayer() string {
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

func Play(videoURL string) error {
	player := getPlayer()
	cmd := exec.Command(player, videoURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
