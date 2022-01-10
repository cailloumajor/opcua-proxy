package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/cailloumajor/opcua-centrifugo/internal/opcua"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
	"github.com/peterbourgon/ff"
)

func usageFor(fs *flag.FlagSet, out io.Writer) func() {
	return func() {
		fmt.Fprintln(out, "USAGE")
		fmt.Fprintf(out, "  %s [options]\n", fs.Name())
		fmt.Fprintln(out)
		fmt.Fprintln(out, "OPTIONS")

		tw := tabwriter.NewWriter(out, 0, 2, 2, ' ', 0)
		fmt.Fprintf(tw, "  Flag\tEnv Var\tDescription\n")
		fs.VisitAll(func(f *flag.Flag) {
			var envVar string
			if f.Name != "debug" {
				envVar = strings.Replace(strings.ToUpper(f.Name), "-", "_", -1)
			}
			var defValue string
			if f.DefValue != "" {
				defValue = fmt.Sprintf(" (default: %s)", f.DefValue)
			}
			fmt.Fprintf(tw, "  -%s\t%s\t%s%s\n", f.Name, envVar, f.Usage, defValue)
		})
		if err := tw.Flush(); err != nil {
			panic(err)
		}
	}
}

func main() {
	var opcuaConfig opcua.Config

	fs := flag.NewFlagSet("opcua-centrifugo", flag.ExitOnError)
	fs.StringVar(
		&opcuaConfig.ServerURL,
		"opcua-server-url",
		"opc.tcp://127.0.0.1:4840",
		"OPC-UA server endpoint URL",
	)
	fs.StringVar(
		&opcuaConfig.User,
		"opcua-user",
		"",
		"user name for OPC-UA authentication (optional)",
	)
	fs.StringVar(
		&opcuaConfig.Password,
		"opcua-password",
		"",
		"password for OPC-UA authentication (optional)",
	)
	fs.StringVar(
		&opcuaConfig.CertFile,
		"opcua-cert-file",
		"",
		"certificate file path for OPC-UA secure channel (optional)",
	)
	fs.StringVar(
		&opcuaConfig.KeyFile,
		"opcua-key-file",
		"",
		"private key file path for OPC-UA secure channel (optional)",
	)
	debug := fs.Bool("debug", false, "log debug information")
	fs.Usage = usageFor(fs, os.Stderr)

	_ = ff.Parse(
		fs,
		os.Args[1:],
		ff.WithEnvVarNoPrefix(),
	)

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	loglevel := level.AllowInfo()
	if *debug {
		loglevel = level.AllowDebug()
	}
	logger = level.NewFilter(logger, loglevel)

	var g run.Group
	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	level.Info(logger).Log("exit", g.Run())
}
