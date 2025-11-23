package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ftl/tetra-cli/pkg/cli"
	"github.com/ftl/tetra-cli/pkg/radio"
	"github.com/spf13/cobra"

	"github.com/ftl/tetra-mess/pkg/demo"
)

const defaultCommandTimeout = 5 * time.Second

var version = "development"

var rootCmd = &cobra.Command{
	Use:     "tetra-mess",
	Version: version,
	Short:   "Measure TETRA signal strength and quality.",
}

func init() {
	cli.InitDefaultTetraFlags(rootCmd, defaultCommandTimeout)
}

func Execute() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fatal(err)
	}
}

func runWithPEI(run func(context.Context, radio.PEI, *cobra.Command, []string)) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		if strings.ToLower(cli.DefaultTetraFlags.Device) == "demo" {
			run(cmd.Context(), demo.NewDemo(), cmd, args)
			return
		}
		cli.RunWithPEI(run, fatal)(cmd, args)
	}
}

func fatal(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func fatalf(format string, args ...any) {
	fatal(fmt.Errorf(format, args...))
}

func logErrorf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
