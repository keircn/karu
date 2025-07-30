package cmd

import (
	"fmt"

	"github.com/keircn/karu/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long:  `View and modify Karu configuration settings for player preferences and download directories.`,
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		key := args[0]
		value := cfg.Get(key)
		if value == "" {
			fmt.Printf("Unknown config key: %s\n", key)
			return
		}
		fmt.Printf("%s = %s\n", key, value)
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		key := args[0]
		value := args[1]

		if err := cfg.Set(key, value); err != nil {
			fmt.Printf("Error setting config: %v\n", err)
			return
		}

		fmt.Printf("Set %s = %s\n", key, value)
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		keys := []string{"player", "player_args", "quality", "download_dir", "auto_play_next", "show_subtitles"}
		fmt.Println("Current configuration:")
		for _, key := range keys {
			value := cfg.Get(key)
			fmt.Printf("  %s = %s\n", key, value)
		}
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	Run: func(cmd *cobra.Command, args []string) {
		defaultCfg := config.DefaultConfig
		if err := config.Save(&defaultCfg); err != nil {
			fmt.Printf("Error resetting config: %v\n", err)
			return
		}
		fmt.Println("Configuration reset to defaults")
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := config.GetConfigPath()
		if err != nil {
			fmt.Printf("Error getting config path: %v\n", err)
			return
		}
		fmt.Println(path)
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configResetCmd)
	configCmd.AddCommand(configPathCmd)
	rootCmd.AddCommand(configCmd)
}
