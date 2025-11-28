package demo

import (
	"context"
	"fmt"
	"time"
)

type DemoPEI struct {
	closed             bool
	disconnectCallback func()
}

func NewDemo() *DemoPEI {
	return &DemoPEI{}
}

func (p *DemoPEI) Close() {
	p.closed = true
	if p.disconnectCallback != nil {
		p.disconnectCallback()
	}
}

func (p *DemoPEI) Closed() bool {
	return p.closed
}

func (p *DemoPEI) WaitUntilClosed(ctx context.Context) {
	// no-op
}

func (p *DemoPEI) OnDisconnect(callback func()) {
	p.disconnectCallback = callback
}

func (p *DemoPEI) AddIndication(prefix string, trailingLines int, handler func(lines []string)) error {
	return nil
}

func (p *DemoPEI) ClearSyntaxErrors(ctx context.Context) error {
	return nil
}

func (p *DemoPEI) Request(ctx context.Context, request string) ([]string, error) {
	return p.AT(ctx, request)
}

func (p *DemoPEI) AT(ctx context.Context, request string) ([]string, error) {
	// log.Printf("AT: %q", request)
	switch request {
	case "AT+GPSPOS?":
		return p.currentGPSPosition()
	case "AT+CSQ?":
		return p.currentSignalStrength()
	case "AT+GCLI?":
		return p.currentCellListInfo()
	default:
		return []string{"OK"}, nil
	}

}

func (p *DemoPEI) ATs(ctx context.Context, requests ...string) error {
	for _, request := range requests {
		_, err := p.AT(ctx, request)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *DemoPEI) currentGPSPosition() ([]string, error) {
	// fmt.Fprintf(os.Stderr, "+GPSPOS %s,N: 52_20.9931,E: 013_22.6631,7\n", time.Now().Format("15:04:05"))
	return []string{
		fmt.Sprintf("+GPSPOS: %s,N: 52_20.9931,E: 013_22.6631,7", time.Now().UTC().Format("15:04:05")),
	}, nil
}

func (p *DemoPEI) currentSignalStrength() ([]string, error) {
	return []string{"+CSQ: 26,99"}, nil
}

func (p *DemoPEI) currentCellListInfo() ([]string, error) {
	return []string{
		"+GCLI: 2",
		"12345,caffe,26,43",
		"12346,caffd,14,19",
	}, nil
}
