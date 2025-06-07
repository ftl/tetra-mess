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
	"github.com/ftl/tetra-mess/pkg/quality"
)

var evalFlags = struct {
	name           string
	outputFilename string
}{}

var evalTrackFlags = struct {
	lac          string
	carrier      string
	outputFormat string
}{}

var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Evaluate a signal trace file",
}

var evalTrackCmd = &cobra.Command{
	Use:   "track [tracefile][ tracefile...]",
	Short: "Convert a signal trace file to a track file in the GPX or KML format",
	Long: `Convert a signal trace file to a track file in the GPX or KML format.
If no LAC or carrier is given, the best server will be used for each GPS position.
If no output filename is given, the filename is derived from the trace filename(s).
`,
	Run: runEvalTrack,
}

var evalQualityCmd = &cobra.Command{
	Use:   "quality [tracefile][ tracefile...]",
	Short: "Evaluate the measurements from one or more signal trace files to visualize the coverage and signal quality of 100x100m fields",
	Run:   runEvalQuality,
}

func init() {
	evalCmd.PersistentFlags().StringVar(&evalFlags.outputFilename, "output", "", "output filename")
	evalCmd.PersistentFlags().StringVar(&evalFlags.name, "name", "", "a name for the evaluation result (default: derived from the input filename)")

	evalTrackCmd.Flags().StringVar(&evalTrackFlags.lac, "lac", "", "LAC of a specific base station to filter for (can be given as decimal or hexadecimal value)")
	evalTrackCmd.Flags().StringVar(&evalTrackFlags.carrier, "carrier", "", "carrier of a specific base station to filter for (can be given as decimal or hexadecimal value)")
	evalTrackCmd.Flags().StringVar(&evalTrackFlags.outputFormat, "format", "kml", "output format (gpx, kml)")

	evalCmd.AddCommand(evalTrackCmd)
	evalCmd.AddCommand(evalQualityCmd)
	rootCmd.AddCommand(evalCmd)
}

type trackWriter func(out io.Writer, trackname string, dataPoints []data.DataPoint) error

func runEvalTrack(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Help()
		return
	}

	var filter data.Filter
	var err error
	var value uint32
	switch {
	case evalTrackFlags.lac != "":
		value, err = data.ParseDecOrHex(evalTrackFlags.lac)
		if err == nil {
			filter = data.FilterByLAC(value)
		}
	case evalTrackFlags.carrier != "":
		value, err = data.ParseDecOrHex(evalTrackFlags.carrier)
		if err == nil {
			filter = data.FilterByCarrier(value)
		}
	default:
		filter = data.FilterBestServer()
	}
	if err != nil {
		cmd.PrintErrf("Error parsing filter value: %v\n", err)
		return
	}

	var writeTrack trackWriter
	switch strings.ToLower(evalTrackFlags.outputFormat) {
	case "gpx":
		writeTrack = gpx.WriteDataPointsAsGPX
	case "kml":
		writeTrack = kml.WriteDataPointsAsKML
	default:
		cmd.PrintErrf("Unsupported output format: %s\n", evalTrackFlags.outputFormat)
		return
	}

	for _, inputFilename := range args {
		outputFilename := evalFlags.outputFilename
		if evalFlags.outputFilename == "" {
			outputFilename = outputFilenameFor(inputFilename, evalTrackFlags.outputFormat)
		}

		trackname := evalFlags.name
		if trackname == "" {
			trackname = filepath.Base(inputFilename)
		}

		err := processTrackInputFile(inputFilename, outputFilename, trackname, filter, writeTrack)
		if err != nil {
			cmd.PrintErrf("Error processing input file %s into %s: %v\n", inputFilename, outputFilename, err)
			continue
		}
	}
}

func processTrackInputFile(inputFilename, outputFilename string, trackname string, filter data.Filter, writeTrack trackWriter) error {
	outputFile, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	dataPoints, err := readInputFile(inputFilename)
	if err != nil {
		return err
	}

	dataPoints = filter.Filter(dataPoints)

	return writeTrack(outputFile, trackname, dataPoints)
}

func runEvalQuality(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Help()
		return
	}
	name := evalFlags.name
	outputFilename := evalFlags.outputFilename

	qualityReport := quality.NewQualityReport()
	for _, inputFilename := range args {
		if outputFilename == "" && evalFlags.outputFilename == "" {
			outputFilename = outputFilenameFor(inputFilename, evalTrackFlags.outputFormat)
		}
		if name == "" {
			name = filepath.Base(inputFilename)
		}

		err := processQualityInputFile(inputFilename, qualityReport)
		if err != nil {
			cmd.PrintErrf("Error processing input file %s: %v\n", inputFilename, err)
			continue
		}
	}
	fieldReports := qualityReport.FieldReports()

	outputFile, err := os.Create(outputFilename)
	if err != nil {
		cmd.PrintErrf("Error creating output file %s: %v\n", outputFilename, err)
		return
	}
	defer outputFile.Close()
	kml.WriteFieldReportsAsKML(outputFile, name, fieldReports)
}

func processQualityInputFile(inputFilename string, qualityReport *quality.QualityReport) error {
	dataPoints, err := readInputFile(inputFilename)
	if err != nil {
		return err
	}

	for _, dataPoint := range dataPoints {
		qualityReport.Add(dataPoint)
	}
	return nil
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

func readInputFile(inputFilename string) ([]data.DataPoint, error) {
	file, err := os.Open(inputFilename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return data.ReadDataPoints(file)
}
