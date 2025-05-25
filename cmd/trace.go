package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/ftl/tetra-pei/com"
	"github.com/spf13/cobra"

	"github.com/ftl/tetra-mess/pkg/data"
	"github.com/ftl/tetra-mess/pkg/scanner"
)

const defaultTraceScanInterval = 10 * time.Second

var traceFlags = struct {
	scanInterval   time.Duration
	outputFilename string
	outputFormat   string
	onlyValid      bool
}{}

var traceCmd = &cobra.Command{
	Use:   "trace",
	Short: "Trace the signal strength and the GPS position and save it to a file",
	Run:   runWithRadio(runTrace), // do not use runWithRadioAndTimeout here, because we want to run the command indefinitely
}

func init() {
	traceCmd.Flags().DurationVar(&traceFlags.scanInterval, "scan-interval", defaultTraceScanInterval, "scan interval")
	traceCmd.Flags().StringVar(&traceFlags.outputFilename, "output", "", "output filename")
	traceCmd.Flags().StringVar(&traceFlags.outputFormat, "format", "csv", "output format (csv, json)")
	traceCmd.Flags().BoolVar(&traceFlags.onlyValid, "only-valid", false, "output only valid data points (with GPS position and RSSI/CSNR values)")

	traceCmd.Flags().MarkHidden("output")

	rootCmd.AddCommand(traceCmd)
}

func runTrace(ctx context.Context, radio *com.COM, cmd *cobra.Command, args []string) {
	err := radio.ATs(ctx,
		"ATZ",
		"ATE0",
		"AT+CSCS=8859-1",
	)
	if err != nil {
		fatalf("cannot initilize radio: %v", err)
	}

	var out io.Writer
	if traceFlags.outputFilename != "" {
		file, err := os.Create(traceFlags.outputFilename)
		if err != nil {
			fatalf("cannot create output file %s: %v", traceFlags.outputFilename, err)
		}
		out = file
	} else {
		out = os.Stdout
	}
	format := TraceOutputFormat(traceFlags.outputFormat)
	onlyValid := traceFlags.onlyValid

	closed := make(chan struct{})
	go func() {
		defer close(closed)

		scanTicker := time.NewTicker(traceFlags.scanInterval)
		defer scanTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-scanTicker.C:
				scan(ctx, radio, out, format, onlyValid)
			}
		}
	}()

	select {
	case <-ctx.Done():
	case <-closed:
	}

	closer, ok := out.(io.Closer)
	if ok {
		closer.Close()
	}
}

type TraceOutputFormat string

const (
	CSVOutput  TraceOutputFormat = "csv"
	JSONOutput TraceOutputFormat = "json"
)

func scan(ctx context.Context, radio *com.COM, out io.Writer, format TraceOutputFormat, onlyValid bool) {
	datapoints, err := scanner.ScanSignalAndPosition(ctx, radio)
	if err != nil {
		log.Printf("error scanning signal and position: %v", err) // TODO: write to stderr
		return
	}

	var encoder func(data.DataPoint) string
	switch format {
	case CSVOutput:
		encoder = data.DataPointToCSV
	case JSONOutput:
		encoder = data.DataPointToJSON
	default:
		fatalf("unknown output format: %s", traceFlags.outputFormat)
	}
	for _, dataPoint := range datapoints {
		if onlyValid && !isDataPointValid(dataPoint) {
			continue
		}

		line := encoder(dataPoint)

		_, err := fmt.Fprintln(out, line)
		if err != nil {
			log.Printf("error writing data point: %v", err) // TODO: write to stderr
			return
		}
	}
}

func isDataPointValid(dataPoint data.DataPoint) bool {
	return dataPoint.Satellites > 0 && dataPoint.RSSI != 99
}
