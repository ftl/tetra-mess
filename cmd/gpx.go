package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ftl/tetra-mess/pkg/data"
)

var gpxFlags = struct {
	lac            string
	id             string
	outputFilename string
}{}

var gpxCmd = &cobra.Command{
	Use:   "gpx [tracefile][ tracefile...]",
	Short: "Convert a signal trace file to a GPX track",
	Long: `Convert a signal trace file to a GPX track.
If no LAC or ID is given, the best server will be used for each GPS position.
If no output filename is given, the filename is derived from the trace filename(s).
`,
	Run: runGpx,
}

func init() {
	gpxCmd.Flags().StringVar(&gpxFlags.lac, "lac", "", "LAC of a specific base station to filter for (can be given as decimal or hexadecimal value)")
	gpxCmd.Flags().StringVar(&gpxFlags.outputFilename, "id", "", "ID of a specific base station to filter for (can be given as decimal or hexadecimal value)")
	gpxCmd.Flags().StringVar(&gpxFlags.outputFilename, "output", "", "output filename")

	rootCmd.AddCommand(gpxCmd)
}

func runGpx(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Help()
		return
	}

	for _, inputFilename := range args {
		outputFilename := gpxFlags.outputFilename
		if gpxFlags.outputFilename == "" {
			outputFilename = outputFilenameFor(inputFilename)
		}

		dataPoints, err := readInputFile(inputFilename)
		if err != nil {
			cmd.PrintErrf("Error reading input file %s: %v\n", inputFilename, err)
			continue
		}
		cmd.Printf("read %d data points from %s to %s\n", len(dataPoints), inputFilename, outputFilename)
	}
}

func outputFilenameFor(inputFilename string) string {
	if inputFilename == "" {
		panic("input filename cannot be empty")
	}

	dir := filepath.Dir(inputFilename)
	base := filepath.Base(inputFilename)
	filename := base[:len(base)-len(filepath.Ext(base))]
	return filepath.Join(dir, filename+".gpx")
}

func readInputFile(inputFilename string) ([]data.DataPoint, error) {
	file, err := os.Open(inputFilename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return data.ReadDataPoints(file)
}
