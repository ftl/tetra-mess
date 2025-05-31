package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ftl/tetra-mess/pkg/quality"
	"github.com/ftl/tetra-mess/pkg/scanner"
	"github.com/ftl/tetra-mess/pkg/tui"
	"github.com/ftl/tetra-pei/com"
	"github.com/spf13/cobra"
)

const defaultTUIScanInterval = 10 * time.Second

var tuiFlags = struct {
	scanInterval time.Duration
}{}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Start the TUI to monitor measurement data and control tracing",
	Run:   runWithRadio(runTUI),
}

func init() {
	tuiCmd.Flags().DurationVar(&tuiFlags.scanInterval, "scan-interval", defaultTUIScanInterval, "scan interval")

	rootCmd.AddCommand(tuiCmd)
}

func runTUI(ctx context.Context, radio *com.COM, cmd *cobra.Command, args []string) {
	// radio
	radioData, err := setupRadioForTUI(ctx, radio)
	if err != nil {
		fatalf("cannot setup radio: %v", err)
	}
	defer func() {
		fmt.Println("Closing radio connection...")
		radio.Close()
	}()

	// UI
	mainScreen := tui.NewMainScreen(version)
	// p := tea.NewProgram(mainScreen)
	p := tea.NewProgram(mainScreen, tea.WithAltScreen())

	go func() {
		for rd := range radioData {
			p.Send(rd)
		}
	}()

	_, err = p.Run()
	if err != nil {
		fatalf("error running TUI: %v", err)
	}
	p.Wait()
}

func setupRadioForTUI(ctx context.Context, radio *com.COM) (<-chan tui.RadioData, error) {
	err := radio.ATs(ctx,
		"ATZ",
		"ATE0",
		"AT+CSCS=8859-1",
	)
	if err != nil {
		return nil, fmt.Errorf("cannot initilize radio: %v", err)
	}

	// radio loop
	radioData := make(chan tui.RadioData, 1)
	go func() {
		defer close(radioData)
		defer fmt.Println("Radio loop closed")

		scanTicker := time.NewTicker(traceFlags.scanInterval)
		defer scanTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-scanTicker.C:
				scanForTUI(ctx, radio, radioData)
			}
		}
	}()

	return radioData, nil
}

func scanForTUI(ctx context.Context, radio *com.COM, radioData chan<- tui.RadioData) {
	position, dataPoints, err := scanner.ScanSignalAndPosition(ctx, radio)
	if err != nil {
		log.Printf("error scanning signal and position: %v", err) // TODO: write to stderr
		return
	}
	measurement := quality.Measurement{}
	measurement.Add(dataPoints...)
	radioData <- tui.RadioData{
		Position:    position,
		Measurement: measurement,
	}
}
