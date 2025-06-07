package tui

import (
	"context"
	"io"

	tea "github.com/charmbracelet/bubbletea"
)

type UIThread interface {
	Send(msg tea.Msg)
}

type Logic struct {
	ui UIThread

	do        chan func()
	radioData <-chan RadioData

	traceFile io.WriteCloser
}

func NewLogic(ui UIThread, radioData <-chan RadioData) *Logic {
	return &Logic{
		ui:        ui,
		do:        make(chan func(), 0),
		radioData: radioData,
		traceFile: nil,
	}
}

func (l *Logic) Start(ctx context.Context) {
	go func() {
		l.ui.Send(l)
		for {
			select {
			case <-ctx.Done():
				close(l.do)
				return
			case f := <-l.do:
				f()
			case rd := <-l.radioData:
				l.ui.Send(rd)
			}
		}
	}()
}

func (l *Logic) Do(f func()) {
	l.do <- f
}
