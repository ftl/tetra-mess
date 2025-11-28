package cmd

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ftl/tetra-cli/pkg/cli"
	"github.com/ftl/tetra-cli/pkg/radio"
	"github.com/spf13/cobra"

	"github.com/ftl/tetra-mess/pkg/scanner"
	"github.com/ftl/tetra-mess/pkg/tui"
)

const defaultTUIScanInterval = 10 * time.Second
const defaultTUIScanTimeout = 5 * time.Second

var tuiFlags = struct {
	scanInterval time.Duration
	outputDir    string
	outputFormat string
}{}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Start the TUI to monitor measurement data and control tracing",
	Run:   runWithPEI(runTUI),
}

func init() {
	tuiCmd.Flags().DurationVar(&tuiFlags.scanInterval, "scan-interval", defaultTUIScanInterval, "scan interval")
	tuiCmd.Flags().StringVar(&tuiFlags.outputDir, "output", "", "output directory for trace files")
	tuiCmd.Flags().StringVar(&tuiFlags.outputFormat, "format", "csv", "output format for trace files (csv, json)")

	rootCmd.AddCommand(tuiCmd)
}

func runTUI(ctx context.Context, pei radio.PEI, cmd *cobra.Command, args []string) {
	// UI
	mainScreen := tui.NewMainScreen(version, cli.DefaultTetraFlags.Device)
	ui := tea.NewProgram(mainScreen, tea.WithAltScreen())

	app := tui.NewApp(ui, tuiFlags.outputDir, tuiFlags.outputFormat)
	app.Start(ctx)

	// radio
	radioLog := func(format string, args ...any) {
		timestamp := fmt.Sprintf("[%s] ", time.Now().Format(time.TimeOnly))
		ui.Send(fmt.Sprintf(timestamp+format, args...))
	}
	loop := scanner.NewScanLoop(tuiFlags.scanInterval, defaultTUIScanTimeout, app.RadioData(), radioLog)

	radio, err := radio.Open(ctx, pei, nil)
	if err != nil {
		fatalf("cannot setup radio: %v", err)
	}
	defer func() {
		fmt.Println("Closing radio connection...")
		radio.Close()
	}()
	radio.RunLoop(loop.Run)

	_, err = ui.Run()
	if err != nil {
		fatalf("error running TUI: %v", err)
	}
	ui.Wait()
}
