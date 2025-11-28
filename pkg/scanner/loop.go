package scanner

import (
	"context"
	"time"

	"github.com/ftl/tetra-cli/pkg/radio"

	"github.com/ftl/tetra-mess/pkg/quality"
)

type ScanLoop struct {
	out          chan<- DataPoint
	logger       Logger
	scanInterval time.Duration
	scanTimeout  time.Duration
}

func NewScanLoop(scanInterval, scanTimeout time.Duration, out chan<- DataPoint, logger Logger) *ScanLoop {
	return &ScanLoop{
		out:          out,
		logger:       logger,
		scanInterval: scanInterval,
		scanTimeout:  scanTimeout,
	}
}

func (l *ScanLoop) Run(ctx context.Context, pei radio.PEI) {
	scanTicker := time.NewTicker(l.scanInterval)
	defer scanTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-scanTicker.C:
			select {
			case l.out <- l.scan(ctx, pei):
			default:
				l.log("cannot report scan data point, output channel not ready")
			}
		}
	}
}

func (l *ScanLoop) scan(ctx context.Context, pei radio.PEI) DataPoint {
	ctx, cancel := context.WithTimeout(ctx, l.scanTimeout)
	defer cancel()

	position, dataPoints := ScanSignalAndPosition(ctx, pei, l.log)

	measurement := quality.Measurement{}
	measurement.Add(dataPoints...)

	return DataPoint{
		Position:    position,
		Measurement: measurement,
	}
}

func (l *ScanLoop) log(format string, args ...any) {
	if l.logger == nil {
		return
	}
	l.logger(format, args...)
}
