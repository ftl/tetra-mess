package radio

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/ftl/tetra-cli/pkg/radio"
	"github.com/ftl/tetra-pei/serial"

	"github.com/ftl/tetra-mess/pkg/data"
	"github.com/ftl/tetra-mess/pkg/quality"
	"github.com/ftl/tetra-mess/pkg/scanner"
)

type DataPoint struct {
	Position    data.Position
	Measurement quality.Measurement
}

type State int

const (
	Disconnected State = iota
	Opening
	Running
	Closing
)

type Radio struct {
	state State

	stateOut chan<- State
	dataOut  chan<- DataPoint

	closing chan struct{}
	closed  chan struct{}

	device       string
	pei          radio.PEI
	tracePEIFile io.Writer

	scanInterval time.Duration
	scanTimeout  time.Duration
}

func New(stateOut chan<- State, dataOut chan<- DataPoint, scanInterval time.Duration) *Radio {
	return &Radio{
		state:        Disconnected,
		stateOut:     stateOut,
		dataOut:      dataOut,
		scanInterval: scanInterval,
		scanTimeout:  250 * time.Millisecond,
	}
}

func (r *Radio) Open(ctx context.Context, device string) error {
	pei, err := serial.Open(device)
	if err != nil {
		return err
	}

	r.pei = pei
	r.device = device
	r.transitionTo(Opening)

	err = r.initialize(ctx)
	if err != nil {
		r.transitionTo(Disconnected)
	}
	return err
}

func (f *Radio) OpenWithTrace(ctx context.Context, portName string, tracePEIFile io.Writer) error {
	return fmt.Errorf("not yet implemented")
}

func (f *Radio) Close(ctx context.Context) {
	select {
	case <-f.closing:
		return
	case <-f.closed:
		return
	default:
	}

	f.transitionTo(Closing)

	close(f.closing)
	<-f.closed

	err := f.shutdownRadio(ctx)
	if err != nil {
		f.reportRadioError("cannot shutdown radio properly: %v", err)
	}

	f.transitionTo(Disconnected)
}

func (r *Radio) transitionTo(nextState State) {
	switch nextState {
	case Disconnected:
		r.pei = nil
		r.device = ""
	}

	r.state = nextState
	r.stateOut <- r.state
}

func (r *Radio) initialize(ctx context.Context) error {
	err := r.pei.ClearSyntaxErrors(ctx)
	if err != nil {
		return err
	}

	err = r.pei.ATs(ctx,
		"ATZ",
		"ATE0",
		"AT+CSCS=8859-1",
	)
	if err != nil {
		return err
	}

	r.closing = make(chan struct{})
	r.closed = make(chan struct{})
	go r.scanLoop()

	r.transitionTo(Running)
	return nil
}

func (r *Radio) shutdownRadio(ctx context.Context) error {
	_, err := r.pei.AT(ctx, "ATZ")
	if err != nil {
		return err
	}
	r.pei.Close()
	r.pei.WaitUntilClosed(ctx)
	return nil
}

func (r *Radio) scanLoop() {
	defer close(r.closed)
	defer log.Println("Radio scan loop closed")

	scanTicker := time.NewTicker(r.scanInterval)
	defer scanTicker.Stop()

	for {
		select {
		case <-r.closing:
			return
		case <-scanTicker.C:
			ctx, cancel := context.WithTimeout(context.Background(), r.scanTimeout)
			defer cancel()

			r.scan(ctx)
		}
	}
}

func (r *Radio) scan(ctx context.Context) {
	position, dataPoints := scanner.ScanSignalAndPosition(ctx, r.pei, r.reportRadioError)

	measurement := quality.Measurement{}
	measurement.Add(dataPoints...)

	r.dataOut <- DataPoint{
		Position:    position,
		Measurement: measurement,
	}
}

func (r *Radio) reportRadioError(format string, args ...any) {
	// TODO: forward the error message to the UI to show it properly
}
