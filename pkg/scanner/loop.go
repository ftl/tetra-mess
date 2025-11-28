package scanner

import (
	"context"
	"time"

	"github.com/ftl/tetra-cli/pkg/radio"

	"github.com/ftl/tetra-mess/pkg/quality"
)

type ScanLoop struct {
	dataOut      chan DataPoint
	errorLog     Logger
	scanInterval time.Duration
	scanTimeout  time.Duration
}

func NewScanLoop(scanInterval, scanTimeout time.Duration, logger Logger) *ScanLoop {
	return &ScanLoop{
		dataOut:      make(chan DataPoint, 1),
		errorLog:     logger,
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
			l.scan(ctx, pei)
		}
	}
}

func (l *ScanLoop) scan(ctx context.Context, pei radio.PEI) {
	ctx, cancel := context.WithTimeout(ctx, l.scanTimeout)
	defer cancel()

	position, dataPoints := ScanSignalAndPosition(ctx, pei, l.log)

	measurement := quality.Measurement{}
	measurement.Add(dataPoints...)

	dataPoint := DataPoint{
		Position:    position,
		Measurement: measurement,
	}

	select {
	case l.dataOut <- dataPoint:
	default:
		l.log("cannot report scan data point, output channel not ready")
	}
}

func (l *ScanLoop) log(format string, args ...any) {
	if l.errorLog == nil {
		return
	}
	l.errorLog(format, args...)
}
