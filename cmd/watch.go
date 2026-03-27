package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rushteam/skills-cli/internal/config"
	"github.com/rushteam/skills-cli/internal/watcher"
	"github.com/spf13/cobra"
)

var watchDirection string

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for changes and auto-sync skills",
	Long: `Start a file watcher that monitors skill directories and automatically
syncs changes. Supports three directions:
  push (default): Watch central dir, push to agents
  pull: Watch agent dirs, pull to central
  both: Watch both, sync in both directions`,
	RunE: runWatch,
}

func init() {
	watchCmd.Flags().StringVar(&watchDirection, "direction", "", "Watch direction (push|pull|both)")
}

func runWatch(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	direction := config.NormalizeWatchDirection(cfg.Sync.Watch.Direction)
	if watchDirection != "" {
		direction = config.NormalizeWatchDirection(watchDirection)
	}
	if direction == "" {
		direction = config.WatchCentralToAgents
	}

	showLogo()
	fmt.Println()
	fmt.Println(titleStyle.Render("  Skills Watcher"))
	fmt.Println()

	w, err := watcher.New(cfg, direction)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	if err := w.Start(); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(dimStyle.Render("  Press Ctrl+C to stop"))
	fmt.Println()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	w.Stop()
	fmt.Println()
	fmt.Println(dimStyle.Render("  Watcher stopped"))
	return nil
}
