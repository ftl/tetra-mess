package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ftl/tetra-mess/pkg/data"
	"github.com/ftl/tetra-mess/pkg/radio"
	"github.com/ftl/tetra-mess/pkg/scanner"
)

const defaultTraceScanInterval = 10 * time.Second

var traceFlags = struct {
	scanInterval   time.Duration
	outputFilename string
	onlyValid      bool
}{}

var traceCmd = &cobra.Command{
	Use:   "trace [output filename]",
	Short: "Trace the signal strength and the GPS position and save it to a file",
	Long: `Trace the signal strength and the GPS position and save it to a file
The output file can be in CSV or JSON format, depending on the file extension.`,
	Run: runWithPEI(runTrace), // do not use runWithRadioAndTimeout here, because we want to run the command indefinitely
}

func init() {
	traceCmd.Flags().DurationVar(&traceFlags.scanInterval, "scan-interval", defaultTraceScanInterval, "scan interval")
	traceCmd.Flags().BoolVar(&traceFlags.onlyValid, "only-valid", false, "output only valid data points (with GPS position and RSSI/Cx values)")

	traceCmd.Flags().MarkHidden("output")

	rootCmd.AddCommand(traceCmd)
}

func runTrace(ctx context.Context, pei radio.PEI, cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Help()
		return
	}

	outputFilename := args[0]
	var out io.Writer
	if outputFilename != "" {
		file, err := os.Create(traceFlags.outputFilename)
		if err != nil {
			fatalf("cannot create output file %s: %v", traceFlags.outputFilename, err)
		}
		out = file
	} else {
		out = os.Stdout
	}

	format := TraceOutputFormat(strings.ToLower(filepath.Ext(outputFilename)))
	var encoder func(data.DataPoint) string
	switch format {
	case "csv":
		encoder = data.DataPointToCSV
	case "json":
		encoder = data.DataPointToJSON
	default:
		fatalf("unknown output format: %s", format)
	}

	onlyValid := traceFlags.onlyValid

	err := pei.ATs(ctx,
		"ATZ",
		"ATE0",
		"AT+CSCS=8859-1",
	)
	if err != nil {
		fatalf("cannot initilize radio: %v", err)
	}

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
				scanForTrace(ctx, pei, out, encoder, onlyValid)
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

func scanForTrace(ctx context.Context, pei radio.PEI, out io.Writer, encoder func(data.DataPoint) string, onlyValid bool) {
	_, datapoints := scanner.ScanSignalAndPosition(ctx, pei, logErrorf)
	for _, dataPoint := range datapoints {
		if onlyValid && !dataPoint.IsValid() {
			continue
		}

		line := encoder(dataPoint)

		_, err := fmt.Fprintln(out, line)
		if err != nil {
			logErrorf("error writing data point: %v", err)
			return
		}
	}
}
