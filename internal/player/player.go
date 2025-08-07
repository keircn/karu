package player

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/keircn/karu/internal/config"
)

type PlaybackInfo struct {
	ShowID    string
	ShowTitle string
	Episodes  []string
	Current   string
	VideoURL  string
}

type model struct {
	currentEpisode  int
	episodes        []string
	showID          string
	showTitle       string
	getVideoURLFunc func(showID, episode string) (string, error)
	initialVideoURL string
	status          string
	autoPlay        bool
	showHelp        bool
	quitting        bool
	currentProcess  *exec.Cmd
	program         *tea.Program
}

type playNextMsg struct{}
type playPrevMsg struct{}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			MarginTop(1).
			PaddingLeft(2)

	episodeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))
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
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return formatPlayerError(err, cfg.Player)
	}
	return nil
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

	m := model{
		currentEpisode:  currentIndex,
		episodes:        info.Episodes,
		showID:          info.ShowID,
		showTitle:       info.ShowTitle,
		getVideoURLFunc: getVideoURLFunc,
		initialVideoURL: info.VideoURL,
		status:          fmt.Sprintf("Playing episode %s", info.Episodes[currentIndex]),
		autoPlay:        cfg.AutoPlayNext,
		showHelp:        false,
		quitting:        false,
	}

	p := tea.NewProgram(&m)
	m.program = p

	go func() {
		cmd, err := startVideoProcess(info.VideoURL, cfg)
		if err != nil {
			m.status = fmt.Sprintf("Player error: %v", err)
			return
		}
		m.currentProcess = cmd

		err = cmd.Wait()
		if err != nil {
			return
		}

		m.updateWatchHistory()

		if m.autoPlay && m.currentEpisode < len(m.episodes)-1 && !m.quitting {
			m.program.Send(playNextMsg{})
		} else if !m.quitting {
			m.program.Quit()
		}
	}()
	_, err = p.Run()
	return err
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.killCurrentProcess()
			m.quitting = true
			return m, tea.Quit
		case "n":
			if m.currentEpisode < len(m.episodes)-1 {
				return m, func() tea.Msg { return playNextMsg{} }
			} else {
				m.status = "Already at last episode"
			}
		case "p":
			if m.currentEpisode > 0 {
				return m, func() tea.Msg { return playPrevMsg{} }
			} else {
				m.status = "Already at first episode"
			}
		case "h", "?":
			m.showHelp = !m.showHelp
		}

	case playNextMsg:
		if m.currentEpisode < len(m.episodes)-1 {
			m.killCurrentProcess()

			m.currentEpisode++
			episode := m.episodes[m.currentEpisode]
			m.status = fmt.Sprintf("Loading episode %s...", episode)

			go func() {
				videoURL, err := m.getVideoURLFunc(m.showID, episode)
				if err != nil {
					m.status = fmt.Sprintf("Error loading episode: %v", err)
					return
				}

				cfg, _ := config.Load()
				cmd, err := startVideoProcess(videoURL, cfg)
				if err != nil {
					m.status = fmt.Sprintf("Player error: %v", err)
					return
				}
				m.currentProcess = cmd
				m.status = fmt.Sprintf("Playing episode %s", episode)

				err = cmd.Wait()
				if err != nil && !m.quitting {
					m.status = fmt.Sprintf("Player error: %v", err)
					return
				}

				m.updateWatchHistory()

				if m.autoPlay && m.currentEpisode < len(m.episodes)-1 && !m.quitting {
					m.program.Send(playNextMsg{})
				} else if !m.quitting {
					m.program.Quit()
				}
			}()
		}

	case playPrevMsg:
		if m.currentEpisode > 0 {
			m.killCurrentProcess()

			m.currentEpisode--
			episode := m.episodes[m.currentEpisode]
			m.status = fmt.Sprintf("Loading episode %s...", episode)

			go func() {
				videoURL, err := m.getVideoURLFunc(m.showID, episode)
				if err != nil {
					m.status = fmt.Sprintf("Error loading episode: %v", err)
					return
				}

				cfg, _ := config.Load()
				cmd, err := startVideoProcess(videoURL, cfg)
				if err != nil {
					m.status = fmt.Sprintf("Player error: %v", err)
					return
				}
				m.currentProcess = cmd
				m.status = fmt.Sprintf("Playing episode %s", episode)

				err = cmd.Wait()
				if err != nil && !m.quitting {
					m.status = fmt.Sprintf("Player error: %v", err)
				}

				m.updateWatchHistory()
			}()
		}
	}

	return m, nil
}

func (m *model) updateWatchHistory() {
	if m.showTitle == "" {
		return
	}

	history, err := config.LoadHistory()
	if err != nil {
		return
	}

	episodeNum := m.currentEpisode + 1
	history.UpdateProgress(m.showTitle, episodeNum)
}

func (m *model) View() string {
	if m.quitting {
		return titleStyle.Render("Goodbye!")
	}

	title := titleStyle.Render("Karu Video Player")

	currentEp := "None"
	if m.currentEpisode < len(m.episodes) {
		currentEp = m.episodes[m.currentEpisode]
	}

	episode := fmt.Sprintf("Episode: %s (%d/%d)",
		episodeStyle.Render(currentEp),
		m.currentEpisode+1,
		len(m.episodes))

	status := statusStyle.Render(m.status)

	autoPlayStatus := "disabled"
	if m.autoPlay {
		autoPlayStatus = "enabled"
	}

	progress := ""
	if m.showTitle != "" {
		history, err := config.LoadHistory()
		if err == nil {
			if lastWatched, exists := history.GetProgress(m.showTitle); exists {
				completion := history.GetCompletionPercentage(m.showTitle)
				progress = fmt.Sprintf("Progress: %d/%d episodes (%.1f%% complete)",
					lastWatched, len(m.episodes), completion)
			}
		}
	}

	controls := fmt.Sprintf("Auto-play: %s", autoPlayStatus)

	help := helpStyle.Render("Press 'h' for help")
	if m.showHelp {
		help = helpStyle.Render(`Controls:
  q       - Quit Karu
  n       - Next episode  
  p       - Previous episode
  h/?     - Toggle this help
  ctrl+c  - Force quit`)
	}

	if progress != "" {
		return fmt.Sprintf("%s\n\n%s\n%s\n%s\n%s\n%s\n",
			title, episode, status, progress, controls, help)
	}

	return fmt.Sprintf("%s\n\n%s\n%s\n%s\n%s\n",
		title, episode, status, controls, help)
}

func (m *model) killCurrentProcess() {
	if m.currentProcess != nil && m.currentProcess.Process != nil {
		m.currentProcess.Process.Kill()
		m.currentProcess = nil
	}
}

func startVideoProcess(videoURL string, cfg *config.Config) (*exec.Cmd, error) {
	args := []string{videoURL}
	if cfg.PlayerArgs != "" {
		playerArgs := strings.Fields(cfg.PlayerArgs)
		args = append(playerArgs, args...)
	}

	cmd := exec.Command(cfg.Player, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, formatPlayerError(err, cfg.Player)
	}
	return cmd, nil
}

func formatPlayerError(err error, player string) error {
	if strings.Contains(err.Error(), "executable file not found") ||
		strings.Contains(err.Error(), "no such file or directory") {
		return fmt.Errorf("video player '%s' not found\n\nTo fix this:\n• Install %s: %s\n• Or configure a different player: karu config set player <path-to-player>",
			player, player, getInstallInstructions(player))
	}
	if strings.Contains(err.Error(), "permission denied") {
		return fmt.Errorf("permission denied when running '%s'\n\nTo fix this:\n• Make sure %s is executable: chmod +x %s\n• Or configure a different player: karu config set player <path-to-player>",
			player, player, player)
	}
	return fmt.Errorf("failed to start video player '%s': %v\n\nTry:\n• karu config set player <path-to-different-player>\n• Check if %s is working: %s --version",
		player, err, player, player)
}

func getInstallInstructions(player string) string {
	switch player {
	case "mpv":
		switch runtime.GOOS {
		case "darwin":
			return "brew install mpv"
		case "linux":
			return "sudo apt install mpv (Ubuntu/Debian) or sudo pacman -S mpv (Arch) or sudo dnf install mpv (Fedora)"
		case "windows":
			return "Download from https://mpv.io/installation/ or use winget install mpv"
		default:
			return "Visit https://mpv.io/installation/"
		}
	case "iina":
		if runtime.GOOS == "darwin" {
			return "brew install --cask iina or download from https://iina.io/"
		}
		return "IINA is only available on macOS"
	case "vlc":
		switch runtime.GOOS {
		case "darwin":
			return "brew install --cask vlc"
		case "linux":
			return "sudo apt install vlc (Ubuntu/Debian) or sudo pacman -S vlc (Arch)"
		case "windows":
			return "Download from https://www.videolan.org/vlc/"
		default:
			return "Visit https://www.videolan.org/vlc/"
		}
	default:
		return fmt.Sprintf("Check documentation for %s installation", player)
	}
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
