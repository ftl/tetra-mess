package radio

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/ftl/tetra-cli/pkg/radio"
	"github.com/ftl/tetra-pei/serial"
)

type Initializer interface {
	Initialize(context.Context, radio.PEI) error
}

type LoopFunc func(context.Context, radio.PEI)

type Radio struct {
	device         string
	pei            radio.PEI
	tracePEIWriter io.Writer

	loopCtx    context.Context
	loopCancel context.CancelFunc
	loopGroup  *sync.WaitGroup

	scanInterval time.Duration
	scanTimeout  time.Duration
}

func Open(ctx context.Context, device string, initializer Initializer) (*Radio, error) {
	opener := func() (radio.PEI, error) {
		return serial.Open(device)
	}
	return open(ctx, device, initializer, opener)
}

func OpenWithTrace(ctx context.Context, device string, initializer Initializer, tracePEIWriter io.Writer) (*Radio, error) {
	opener := func() (radio.PEI, error) {
		return serial.OpenWithTrace(device, tracePEIWriter)
	}
	return open(ctx, device, initializer, opener)
}

func open(ctx context.Context, device string, initializer Initializer, opener func() (radio.PEI, error)) (*Radio, error) {
	pei, err := opener()
	if err != nil {
		return nil, err
	}

	loopCtx, loopCancel := context.WithCancel(context.Background())

	result := &Radio{
		device: device,
		pei:    pei,

		loopCtx:    loopCtx,
		loopCancel: loopCancel,
		loopGroup:  new(sync.WaitGroup),
	}
	err = result.initialize(ctx, initializer)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Radio) initialize(ctx context.Context, initializer Initializer) error {
	err := r.pei.ClearSyntaxErrors(ctx)
	if err != nil {
		return err
	}

	// initialize the PEI
	err = r.pei.ATs(ctx,
		"ATZ",
		"ATE0",
		"AT+CSCS=8859-1",
	)
	if err != nil {
		return err
	}

	if initializer == nil {
		return nil
	}
	return initializer.Initialize(ctx, r.pei)
}

func (r *Radio) Close(ctx context.Context) error {
	if !r.Connected() {
		return nil
	}

	// stop the running loops and wait until they are stopped
	r.loopCancel()
	r.loopGroup.Wait()

	// reset the PEI to defaults
	_, err := r.pei.AT(ctx, "ATZ")
	if err != nil {
		return err
	}

	r.pei.Close()
	r.pei.WaitUntilClosed(ctx)
	r.pei = nil

	return nil
}

func (r *Radio) Connected() bool {
	return r.pei != nil && !r.pei.Closed()
}

func (r *Radio) AddIndication(prefix string, trailingLines int, handler func(lines []string)) error {
	return r.pei.AddIndication(prefix, trailingLines, handler)
}

func (r *Radio) Request(ctx context.Context, request string) ([]string, error) {
	return r.Request(ctx, request)
}

func (r *Radio) AT(ctx context.Context, request string) ([]string, error) {
	return r.AT(ctx, request)
}

func (r *Radio) ATs(ctx context.Context, requests ...string) error {
	return r.ATs(ctx, requests...)
}

func (r *Radio) RunLoop(loop LoopFunc) {
	if !r.Connected() {
		return
	}

	r.loopGroup.Go(func() {
		loop(r.loopCtx, r.pei)
	})
}
