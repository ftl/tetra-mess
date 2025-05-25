package cmd

import "github.com/spf13/cobra"

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
}
