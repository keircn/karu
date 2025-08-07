package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "karu",
	Short: "A fast and pretty CLI for watching anime",
	Long:  `Karu is a command-line interface for discovering and watching anime with interactive browsing and seamless video playback.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	setupSignalHandling()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setupSignalHandling() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nReceived interrupt signal. Cleaning up...")
		cleanupAndExit()
	}()
}

func cleanupAndExit() {
	os.Exit(0)
}
