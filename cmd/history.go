package cmd

import (
	"fmt"
	"strconv"

	"github.com/keircn/karu/internal/config"
	"github.com/keircn/karu/internal/ui"
	"github.com/keircn/karu/internal/workflow"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history [subcommand]",
	Short: "Manage search history",
	Long:  `View, search, and manage your anime search history.`,
	Run: func(cmd *cobra.Command, args []string) {
		selection, err := workflow.GetAnimeSelectionFromHistory()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("You chose: %s\n", selection.Anime.Title)

		episode, err := ui.SelectEpisode(selection.Episodes, selection.Anime.Title)
		if err != nil {
			fmt.Printf("Error selecting episode: %v\n", err)
			return
		}

		if episode != nil {
			fmt.Printf("You chose episode: %s\n", *episode)

			history, _ := config.LoadHistory()
			if history != nil {
				episodeNum, _ := strconv.Atoi(*episode)
				history.UpdateProgress(selection.Anime.Title, episodeNum)
			}
		}
	},
}

var historyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all history entries",
	Run: func(cmd *cobra.Command, args []string) {
		history, err := config.LoadHistory()
		if err != nil {
			fmt.Printf("Error loading history: %v\n", err)
			return
		}

		entries := history.GetRecent(0)
		if len(entries) == 0 {
			fmt.Println("No history entries found.")
			return
		}

		fmt.Println("Search History:")
		fmt.Println("===============")
		for i, entry := range entries {
			progress := ""
			if entry.TotalEps > 0 {
				progress = fmt.Sprintf(" (%d/%d)", entry.LastWatched, entry.TotalEps)
			}
			fmt.Printf("%d. %s%s\n", i+1, entry.Title, progress)
			fmt.Printf("   Query: %s\n", entry.Query)
			fmt.Printf("   Watched %d times • %s\n\n", entry.AccessCount, entry.Timestamp.Format("Jan 2, 2006"))
		}
	},
}

var historyRecentCmd = &cobra.Command{
	Use:   "recent [limit]",
	Short: "Show recent history entries",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		limit := 10
		if len(args) > 0 {
			if l, err := strconv.Atoi(args[0]); err == nil && l > 0 {
				limit = l
			}
		}

		history, err := config.LoadHistory()
		if err != nil {
			fmt.Printf("Error loading history: %v\n", err)
			return
		}

		entries := history.GetRecent(limit)
		if len(entries) == 0 {
			fmt.Println("No recent history entries found.")
			return
		}

		fmt.Printf("Recent %d entries:\n", len(entries))
		fmt.Println("==================")
		for i, entry := range entries {
			progress := ""
			if entry.TotalEps > 0 {
				progress = fmt.Sprintf(" (%d/%d)", entry.LastWatched, entry.TotalEps)
			}
			fmt.Printf("%d. %s%s\n", i+1, entry.Title, progress)
			fmt.Printf("   %s\n\n", entry.Timestamp.Format("Jan 2, 2006 15:04"))
		}
	},
}

var historyPopularCmd = &cobra.Command{
	Use:   "popular [limit]",
	Short: "Show most watched anime from history",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		limit := 10
		if len(args) > 0 {
			if l, err := strconv.Atoi(args[0]); err == nil && l > 0 {
				limit = l
			}
		}

		history, err := config.LoadHistory()
		if err != nil {
			fmt.Printf("Error loading history: %v\n", err)
			return
		}

		entries := history.GetMostWatched(limit)
		if len(entries) == 0 {
			fmt.Println("No history entries found.")
			return
		}

		fmt.Printf("Most watched (%d entries):\n", len(entries))
		fmt.Println("===========================")
		for i, entry := range entries {
			progress := ""
			if entry.TotalEps > 0 {
				progress = fmt.Sprintf(" (%d/%d)", entry.LastWatched, entry.TotalEps)
			}
			fmt.Printf("%d. %s%s\n", i+1, entry.Title, progress)
			fmt.Printf("   Watched %d times\n\n", entry.AccessCount)
		}
	},
}

var historySearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search through history entries",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]

		history, err := config.LoadHistory()
		if err != nil {
			fmt.Printf("Error loading history: %v\n", err)
			return
		}

		entries := history.Search(query)
		if len(entries) == 0 {
			fmt.Printf("No history entries found matching '%s'.\n", query)
			return
		}

		fmt.Printf("Search results for '%s' (%d entries):\n", query, len(entries))
		fmt.Println("=====================================")
		for i, entry := range entries {
			progress := ""
			if entry.TotalEps > 0 {
				progress = fmt.Sprintf(" (%d/%d)", entry.LastWatched, entry.TotalEps)
			}
			fmt.Printf("%d. %s%s\n", i+1, entry.Title, progress)
			fmt.Printf("   Query: %s\n", entry.Query)
			fmt.Printf("   %s\n\n", entry.Timestamp.Format("Jan 2, 2006"))
		}
	},
}

var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all history entries",
	Run: func(cmd *cobra.Command, args []string) {
		history, err := config.LoadHistory()
		if err != nil {
			fmt.Printf("Error loading history: %v\n", err)
			return
		}

		if err := history.Clear(); err != nil {
			fmt.Printf("Error clearing history: %v\n", err)
			return
		}

		fmt.Println("History cleared successfully.")
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)

	historyCmd.AddCommand(historyListCmd)
	historyCmd.AddCommand(historyRecentCmd)
	historyCmd.AddCommand(historyPopularCmd)
	historyCmd.AddCommand(historySearchCmd)
	historyCmd.AddCommand(historyClearCmd)
}
