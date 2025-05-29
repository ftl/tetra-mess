package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ftl/tetra-mess/pkg/data"
	"github.com/ftl/tetra-mess/pkg/gpx"
	"github.com/ftl/tetra-mess/pkg/kml"
)

var evalFlags = struct {
	lac            string
	id             string
	outputFilename string
	outputFormat   string
}{}

var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Evaluate a signal trace file",
}

var evalTrackCmd = &cobra.Command{
	Use:   "track [tracefile][ tracefile...]",
	Short: "Convert a signal trace file to a track file in the GPX or KML format",
	Long: `Convert a signal trace file to a track file in the GPX or KML format.
If no LAC or ID is given, the best server will be used for each GPS position.
If no output filename is given, the filename is derived from the trace filename(s).
`,
	Run: runEvalTrack,
}

func init() {
	evalCmd.PersistentFlags().StringVar(&evalFlags.lac, "lac", "", "LAC of a specific base station to filter for (can be given as decimal or hexadecimal value)")
	evalCmd.PersistentFlags().StringVar(&evalFlags.id, "id", "", "ID of a specific base station to filter for (can be given as decimal or hexadecimal value)")
	evalCmd.PersistentFlags().StringVar(&evalFlags.outputFilename, "output", "", "output filename")
	evalCmd.PersistentFlags().StringVar(&evalFlags.outputFormat, "format", "kml", "output format (gpx, kml)")

	evalCmd.AddCommand(evalTrackCmd)
	rootCmd.AddCommand(evalCmd)
}

type trackWriter func(out io.Writer, name string, dataPoints []data.DataPoint) error

func runEvalTrack(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Help()
		return
	}

	var filter data.Filter
	var err error
	var value uint32
	switch {
	case evalFlags.lac != "":
		value, err = data.ParseDecOrHex(evalFlags.lac)
		if err == nil {
			filter = data.FilterByLAC(value)
		}
	case evalFlags.id != "":
		value, err = data.ParseDecOrHex(evalFlags.id)
		if err == nil {
			filter = data.FilterByID(value)
		}
	default:
		filter = data.FilterBestServer()
	}
	if err != nil {
		cmd.PrintErrf("Error parsing filter value: %v\n", err)
		return
	}

	var writeTrack trackWriter
	switch strings.ToLower(evalFlags.outputFormat) {
	case "gpx":
		writeTrack = gpx.WriteDataPointsAsGPX
	case "kml":
		writeTrack = kml.WriteDataPointsAsKML
	default:
		cmd.PrintErrf("Unsupported output format: %s\n", evalFlags.outputFormat)
		return
	}

	for _, inputFilename := range args {
		outputFilename := evalFlags.outputFilename
		if evalFlags.outputFilename == "" {
			outputFilename = outputFilenameFor(inputFilename, evalFlags.outputFormat)
		}

		err := processInputFile(inputFilename, outputFilename, filter, writeTrack)
		if err != nil {
			cmd.PrintErrf("Error processing input file %s into %s: %v\n", inputFilename, outputFilename, err)
			continue
		}
	}
}

func outputFilenameFor(inputFilename string, formatExtension string) string {
	if inputFilename == "" {
		panic("input filename cannot be empty")
	}

	dir := filepath.Dir(inputFilename)
	base := filepath.Base(inputFilename)
	filename := base[:len(base)-len(filepath.Ext(base))]
	return filepath.Join(dir, filename+"."+formatExtension)
}

func processInputFile(inputFilename, outputFilename string, filter data.Filter, writeTrack trackWriter) error {
	outputFile, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	dataPoints, err := readInputFile(inputFilename)
	if err != nil {
		return err
	}
	if len(dataPoints) == 0 {
		return nil // nothing to do
	}

	dataPoints = filter.Filter(dataPoints)

	return writeTrack(outputFile, outputFilename, dataPoints)
}

func readInputFile(inputFilename string) ([]data.DataPoint, error) {
	file, err := os.Open(inputFilename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return data.ReadDataPoints(file)
}
