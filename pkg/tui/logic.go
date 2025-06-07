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

type UIThread interface {
	Send(msg tea.Msg)
}

type Logic struct {
	ui UIThread

	do        chan func() error
	radioData <-chan RadioData

	outputDir    string
	outputFormat string
	traceFile    io.WriteCloser
}

func NewLogic(ui UIThread, radioData <-chan RadioData, outputDir, outputFormat string) *Logic {
	return &Logic{
		ui:           ui,
		do:           make(chan func() error, 0),
		radioData:    radioData,
		outputDir:    outputDir,
		outputFormat: strings.ToLower(outputFormat),
		traceFile:    nil,
	}
}

func (l *Logic) Start(ctx context.Context) {
	go func() {
		defer l.stopTrace()

		l.ui.Send(l)
		for {
			select {
			case <-ctx.Done():
				close(l.do)
				return
			case f := <-l.do:
				err := f()
				if err != nil {
					l.ui.Send(err)
				}
			case rd := <-l.radioData:
				l.traceRadioData(rd)
				l.ui.Send(rd)
			}
		}
	}()
}

func (l *Logic) Do(f func() error) {
	l.do <- f
}

func (l *Logic) ToggleTrace() tea.Msg {
	var f func() error
	if l.traceFile == nil {
		f = l.startTrace
	} else {
		f = l.stopTrace
	}

	l.Do(f)
	return nil
}

// Methods beyond this point MUST ONLY be called from within the goroutine!

func (l *Logic) showMessage(format string, args ...any) {
	l.ui.Send(fmt.Sprintf(format, args...))
}

func (l *Logic) clearMessage() {
	l.ui.Send("")
}

func (l *Logic) traceRadioData(rd RadioData) {
	if l.traceFile == nil {
		return
	}

	var encoder func(data.DataPoint) string
	switch l.outputFormat {
	case "csv":
		encoder = data.DataPointToCSV
	case "json":
		encoder = data.DataPointToJSON
	default:
		l.showMessage("unknown output format: %s", l.outputFormat)
	}

	for _, dataPoint := range rd.Measurement.DataPoints {
		// TODO: add support to trace only valid data points to the TUI
		// if onlyValid && !dataPoint.IsValid() {
		// 	continue
		// }

		line := encoder(dataPoint)

		_, err := fmt.Fprintln(l.traceFile, line)
		if err != nil {
			l.showMessage("error writing data point: %v", err)
			return
		}
	}

}

func (l *Logic) startTrace() error {
	if l.traceFile != nil {
		return nil
	}

	filename := l.newTraceFilename()
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("cannot create trace file: %w", err)
	}
	l.traceFile = file

	l.showMessage("tracing started: %s", filename)

	return nil
}

func (l *Logic) newTraceFilename() string {
	filename := fmt.Sprintf("trace-%s.%s", time.Now().Format("20060102T150405"), l.outputFormat)
	return filepath.Join(l.outputDir, filename)
}

func (l *Logic) stopTrace() error {
	if l.traceFile == nil {
		return nil
	}

	err := l.traceFile.Close()
	l.traceFile = nil

	if err != nil {
		return fmt.Errorf("cannot close the trace file: %w", err)
	}
	l.showMessage("tracing stopped")
	return nil
}
