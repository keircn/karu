package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type HistoryEntry struct {
	ID          string    `json:"id"`
	Query       string    `json:"query"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	LastWatched int       `json:"last_watched"`
	TotalEps    int       `json:"total_episodes"`
	Timestamp   time.Time `json:"timestamp"`
	AccessCount int       `json:"access_count"`
}

type History struct {
	Entries    []HistoryEntry `json:"entries"`
	MaxEntries int            `json:"max_entries"`
}

var DefaultHistory = History{
	Entries:    []HistoryEntry{},
	MaxEntries: 50,
}

func GetHistoryPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	karuConfigDir := filepath.Join(configDir, "karu")
	if err := os.MkdirAll(karuConfigDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(karuConfigDir, "history.json"), nil
}

func LoadHistory() (*History, error) {
	historyPath, err := GetHistoryPath()
	if err != nil {
		return &DefaultHistory, nil
	}

	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		history := DefaultHistory
		if err := SaveHistory(&history); err != nil {
			return &DefaultHistory, nil
		}
		return &history, nil
	}

	data, err := os.ReadFile(historyPath)
	if err != nil {
		return &DefaultHistory, err
	}

	history := DefaultHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return &DefaultHistory, err
	}

	if history.MaxEntries <= 0 {
		history.MaxEntries = DefaultHistory.MaxEntries
	}

	return &history, nil
}

func SaveHistory(history *History) error {
	historyPath, err := GetHistoryPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyPath, data, 0644)
}

func (h *History) AddEntry(query, title, url string, totalEps int) error {
	now := time.Now()

	for i, entry := range h.Entries {
		if entry.Title == title {
			h.Entries[i].Query = query
			h.Entries[i].URL = url
			h.Entries[i].TotalEps = totalEps
			h.Entries[i].Timestamp = now
			h.Entries[i].AccessCount++
			return SaveHistory(h)
		}
	}

	newEntry := HistoryEntry{
		ID:          generateID(title),
		Query:       query,
		Title:       title,
		URL:         url,
		LastWatched: 1,
		TotalEps:    totalEps,
		Timestamp:   now,
		AccessCount: 1,
	}

	h.Entries = append(h.Entries, newEntry)

	if len(h.Entries) > h.MaxEntries {
		sort.Slice(h.Entries, func(i, j int) bool {
			return h.Entries[i].Timestamp.After(h.Entries[j].Timestamp)
		})
		h.Entries = h.Entries[:h.MaxEntries]
	}

	return SaveHistory(h)
}

func (h *History) UpdateProgress(title string, episode int) error {
	for i, entry := range h.Entries {
		if entry.Title == title {
			if episode > h.Entries[i].LastWatched {
				h.Entries[i].LastWatched = episode
			}
			h.Entries[i].Timestamp = time.Now()
			h.Entries[i].AccessCount++
			return SaveHistory(h)
		}
	}
	return nil
}

func (h *History) GetProgress(title string) (int, bool) {
	for _, entry := range h.Entries {
		if entry.Title == title {
			return entry.LastWatched, true
		}
	}
	return 0, false
}

func (h *History) IsWatched(title string, episode int) bool {
	for _, entry := range h.Entries {
		if entry.Title == title {
			return episode <= entry.LastWatched
		}
	}
	return false
}

func (h *History) GetNextEpisode(title string) int {
	for _, entry := range h.Entries {
		if entry.Title == title {
			if entry.LastWatched < entry.TotalEps {
				return entry.LastWatched + 1
			}
			return entry.LastWatched
		}
	}
	return 1
}

func (h *History) GetCompletionPercentage(title string) float64 {
	for _, entry := range h.Entries {
		if entry.Title == title && entry.TotalEps > 0 {
			return float64(entry.LastWatched) / float64(entry.TotalEps) * 100
		}
	}
	return 0
}

func (h *History) RemoveEntry(title string) error {
	for i, entry := range h.Entries {
		if entry.Title == title {
			h.Entries = append(h.Entries[:i], h.Entries[i+1:]...)
			return SaveHistory(h)
		}
	}
	return nil
}

func (h *History) Clear() error {
	h.Entries = []HistoryEntry{}
	return SaveHistory(h)
}

func (h *History) GetRecent(limit int) []HistoryEntry {
	entries := make([]HistoryEntry, len(h.Entries))
	copy(entries, h.Entries)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	if limit > 0 && limit < len(entries) {
		entries = entries[:limit]
	}

	return entries
}

func (h *History) GetMostWatched(limit int) []HistoryEntry {
	entries := make([]HistoryEntry, len(h.Entries))
	copy(entries, h.Entries)

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].AccessCount == entries[j].AccessCount {
			return entries[i].Timestamp.After(entries[j].Timestamp)
		}
		return entries[i].AccessCount > entries[j].AccessCount
	})

	if limit > 0 && limit < len(entries) {
		entries = entries[:limit]
	}

	return entries
}

func (h *History) Search(query string) []HistoryEntry {
	var matches []HistoryEntry
	queryLower := strings.ToLower(query)

	for _, entry := range h.Entries {
		if strings.Contains(strings.ToLower(entry.Title), queryLower) ||
			strings.Contains(strings.ToLower(entry.Query), queryLower) {
			matches = append(matches, entry)
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Timestamp.After(matches[j].Timestamp)
	})

	return matches
}

func generateID(title string) string {
	return strings.ReplaceAll(strings.ToLower(title), " ", "-")
}
