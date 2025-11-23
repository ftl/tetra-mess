package cmd

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ftl/tetra-cli/pkg/cli"
	"github.com/ftl/tetra-cli/pkg/radio"
	"github.com/spf13/cobra"

	"github.com/ftl/tetra-mess/pkg/quality"
	"github.com/ftl/tetra-mess/pkg/scanner"
	"github.com/ftl/tetra-mess/pkg/tui"
)

const defaultTUIScanInterval = 10 * time.Second

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
	err := setupRadio(ctx, pei, tuiFlags.scanInterval, app.RadioData())
	if err != nil {
		fatalf("cannot setup radio: %v", err)
	}
	defer func() {
		fmt.Println("Closing radio connection...")
		pei.Close()
	}()

	_, err = ui.Run()
	if err != nil {
		fatalf("error running TUI: %v", err)
	}
	ui.Wait()
}

func setupRadio(ctx context.Context, pei radio.PEI, scanInterval time.Duration, radioData chan<- tui.RadioData) error {
	err := pei.ATs(ctx,
		"ATZ",
		"ATE0",
		"AT+CSCS=8859-1",
	)
	if err != nil {
		return fmt.Errorf("cannot initilize radio: %v", err)
	}

	// radio loop
	go func() {
		defer fmt.Println("Radio loop closed")

		scanTicker := time.NewTicker(scanInterval)
		defer scanTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-scanTicker.C:
				scanForTUI(ctx, pei, radioData)
			}
		}
	}()

	return nil
}

func scanForTUI(ctx context.Context, pei radio.PEI, radioData chan<- tui.RadioData) {
	position, dataPoints := scanner.ScanSignalAndPosition(ctx, pei, logErrorf) // TODO: forward the error message to the UI to show it properly
	measurement := quality.Measurement{}
	measurement.Add(dataPoints...)
	radioData <- tui.RadioData{
		Position:    position,
		Measurement: measurement,
	}
}
