package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ftl/tetra-mess/pkg/radio"
	"github.com/ftl/tetra-pei/com"
	"github.com/hedhyw/Go-Serial-Detector/pkg/v1/serialdet"
	"github.com/jacobsa/go-serial/serial"
	"github.com/spf13/cobra"
)

const defaultCommandTimeout = 5 * time.Second

var version = "development"

var rootFlags = struct {
	device           string
	commandTimeout   time.Duration
	tracePEIFilename string
}{}

var rootCmd = &cobra.Command{
	Use:     "tetra-mess",
	Version: version,
	Short:   "Measure TETRA signal strength and quality.",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&rootFlags.device, "device", "", "serial communication device (leave empty for auto detection)")
	rootCmd.PersistentFlags().DurationVar(&rootFlags.commandTimeout, "commandTimeout", defaultCommandTimeout, "timeout for commands")
	rootCmd.PersistentFlags().StringVar(&rootFlags.tracePEIFilename, "trace-pei", "", "filename for tracing the PEI communication")
	rootCmd.PersistentFlags().MarkHidden("trace-pei")
}

func Execute() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fatal(err)
	}
}

func runWithPEIAndTimeout(run func(context.Context, radio.PEI, *cobra.Command, []string)) func(*cobra.Command, []string) {
	return runWithPEI(func(ctx context.Context, pei radio.PEI, cmd *cobra.Command, args []string) {
		cmdCtx, cancel := context.WithTimeout(ctx, rootFlags.commandTimeout)
		defer cancel()
		run(cmdCtx, pei, cmd, args)
	})
}

func runWithPEI(run func(context.Context, radio.PEI, *cobra.Command, []string)) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		var err error
		rootCtx := cmd.Context()

		var tracePEIFile *os.File
		if rootFlags.tracePEIFilename != "" {
			tracePEIFile, err = os.OpenFile(rootFlags.tracePEIFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fatalf("cannot access PEI trace file: %v", err)
			}
			defer tracePEIFile.Close()
		}

		portName, err := getRadioPortName(rootFlags.device)
		if err != nil {
			fatal(err)
		}
		rootFlags.device = portName

		var pei radio.PEI
		var device io.ReadWriteCloser
		if portName == "demo" {
			pei = radio.NewDemo()
		} else {
			pei, device = setupPEI(rootCtx, portName, tracePEIFile)
		}
		defer func() {
			if device != nil {
				device.Close()
			}
		}()

		run(rootCtx, pei, cmd, args)

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), rootFlags.commandTimeout)
		defer cancelShutdown()
		pei.AT(shutdownCtx, "ATZ")
		pei.Close()
		pei.WaitUntilClosed(shutdownCtx)
	}
}

func getRadioPortName(deviceFlag string) (string, error) {
	if deviceFlag != "" && strings.ToLower(deviceFlag) != "auto" {
		return deviceFlag, nil
	}

	devices, err := serialdet.List()
	if err != nil {
		return "", err
	}

	for _, device := range devices {
		description := strings.ToLower(device.Description())
		if strings.Contains(description, "tetra_pei_interface") {
			return device.Path(), nil
		}
	}

	return "", fmt.Errorf("no active PEI interface found, use the --device parameter to provide the serial communication device")
}

func setupPEI(ctx context.Context, portName string, tracePEIFile io.Writer) (radio.PEI, io.ReadWriteCloser) {
	portConfig := serial.OpenOptions{
		PortName:              portName,
		BaudRate:              38400,
		DataBits:              8,
		StopBits:              1,
		ParityMode:            serial.PARITY_NONE,
		RTSCTSFlowControl:     true,
		MinimumReadSize:       4,
		InterCharacterTimeout: 100,
	}
	device, err := serial.Open(portConfig)
	if err != nil {
		fatal(err)
	}

	var pei radio.PEI
	if tracePEIFile != nil {
		pei = com.NewWithTrace(device, tracePEIFile)
	} else {
		pei = com.New(device)
	}
	err = pei.ClearSyntaxErrors(ctx)
	if err != nil {
		fatalf("cannot connect to radio: %v", err)
	}

	return pei, device
}

func fatal(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func fatalf(format string, args ...any) {
	fatal(fmt.Errorf(format, args...))
}
