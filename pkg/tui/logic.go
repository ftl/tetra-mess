package tui

import (
	"context"
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
)

type UIThread interface {
	Send(msg tea.Msg)
}

type Logic struct {
	ui UIThread

	do        chan func() error
	radioData <-chan RadioData

	traceFile io.WriteCloser
}

func NewLogic(ui UIThread, radioData <-chan RadioData) *Logic {
	return &Logic{
		ui:        ui,
		do:        make(chan func() error, 0),
		radioData: radioData,
		traceFile: nil,
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

	// TODO: write radioData to the trace file
}

func (l *Logic) startTrace() error {
	if l.traceFile != nil {
		return nil
	}

	// TODO: implement starting the trace
	// - generate filename
	// - open the file
	// - set l.traceFile

	l.showMessage("tracing started")

	return fmt.Errorf("not yet implemented")
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
