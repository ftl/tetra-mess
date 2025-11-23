package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ftl/tetra-mess/pkg/data"
)

type UI interface {
	Send(msg tea.Msg)
}

type App struct {
	ui UI

	do        chan func() error
	radioData chan RadioData

	outputDir    string
	outputFormat string
	traceFile    io.WriteCloser
}

func NewApp(ui UI, outputDir, outputFormat string) *App {
	return &App{
		ui:           ui,
		do:           make(chan func() error),
		radioData:    make(chan RadioData, 1),
		outputDir:    outputDir,
		outputFormat: strings.ToLower(outputFormat),
		traceFile:    nil,
	}
}

func (a *App) Start(ctx context.Context) {
	go func() {
		defer a.stopTrace()

		a.ui.Send(a)
		for {
			select {
			case <-ctx.Done():
				close(a.do)
				close(a.radioData)
				return
			case f := <-a.do:
				err := f()
				if err != nil {
					a.ui.Send(err)
				}
			case rd := <-a.radioData:
				a.traceRadioData(rd)
				a.ui.Send(rd)
			}
		}
	}()
}

func (a *App) RadioData() chan<- RadioData {
	return a.radioData
}

func (a *App) Do(f func() error) {
	a.do <- f
}

func (a *App) ToggleTrace() tea.Msg {
	var f func() error
	if a.traceFile == nil {
		f = a.startTrace
	} else {
		f = a.stopTrace
	}

	a.Do(f)
	return nil
}

// Methods beyond this point MUST ONLY be called from within the goroutine!

func (a *App) showMessage(format string, args ...any) {
	a.ui.Send(fmt.Sprintf(format, args...))
}

func (a *App) sendStatus(filename string, active bool) {
	a.ui.Send(TracingStatus{Filename: filename, Active: active})
}

func (a *App) traceRadioData(rd RadioData) {
	if a.traceFile == nil {
		return
	}

	var encoder func(data.DataPoint) string
	switch a.outputFormat {
	case "csv":
		encoder = data.DataPointToCSV
	case "json":
		encoder = data.DataPointToJSON
	default:
		a.showMessage("unknown output format: %s", a.outputFormat)
	}

	for _, dataPoint := range rd.Measurement.DataPoints {
		// TODO: add support to trace only valid data points to the TUI
		// if onlyValid && !dataPoint.IsValid() {
		// 	continue
		// }

		line := encoder(dataPoint)

		_, err := fmt.Fprintln(a.traceFile, line)
		if err != nil {
			a.showMessage("error writing data point: %v", err)
			return
		}
	}

}

func (a *App) startTrace() error {
	if a.traceFile != nil {
		return nil
	}

	filename := a.newTraceFilename()
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("cannot create trace file: %w", err)
	}
	a.traceFile = file

	a.showMessage("tracing started")
	a.sendStatus(filename, true)

	return nil
}

func (a *App) newTraceFilename() string {
	filename := fmt.Sprintf("trace-%s.%s", time.Now().Format("20060102T150405"), a.outputFormat)
	return filepath.Join(a.outputDir, filename)
}

func (a *App) stopTrace() error {
	if a.traceFile == nil {
		return nil
	}

	err := a.traceFile.Close()
	a.traceFile = nil

	if err != nil {
		return fmt.Errorf("cannot close the trace file: %w", err)
	}
	a.showMessage("tracing stopped")
	a.sendStatus("", false)
	return nil
}
