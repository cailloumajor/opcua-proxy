package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/cailloumajor/opcua-centrifugo/internal/opcua"
	"github.com/cailloumajor/opcua-centrifugo/internal/proxy"
	"github.com/centrifugal/gocent/v3"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/joho/godotenv"
	"github.com/oklog/run"
	"github.com/peterbourgon/ff"
)

const stopTimeout = 2 * time.Second

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

func checkErr(l log.Logger, err error, code int) {
	if err != nil {
		level.Info(l).Log("err", err)
		os.Exit(code)
	}
}

func main() {
	var (
		opcuaConfig            opcua.Config
		opcuaClientTimeout     time.Duration
		opcuaTidyInterval      time.Duration
		proxyListen            string
		centrifugoNamespace    string
		centrifugoClientConfig gocent.Config
	)
	fs := flag.NewFlagSet("opcua-centrifugo", flag.ExitOnError)
	fs.StringVar(&opcuaConfig.ServerURL, "opcua-server-url", "opc.tcp://127.0.0.1:4840", "OPC-UA server endpoint URL")
	fs.StringVar(&opcuaConfig.User, "opcua-user", "", "user name for OPC-UA authentication (optional)")
	fs.StringVar(&opcuaConfig.Password, "opcua-password", "", "password for OPC-UA authentication (optional)")
	fs.StringVar(&opcuaConfig.CertFile, "opcua-cert-file", "", "certificate file path for OPC-UA secure channel (optional)")
	fs.StringVar(&opcuaConfig.KeyFile, "opcua-key-file", "", "private key file path for OPC-UA secure channel (optional)")
	fs.DurationVar(&opcuaClientTimeout, "opcua-client-timeout", 10*time.Second, "timeout for connecting the OPC-UA client")
	fs.DurationVar(&opcuaTidyInterval, "opcua-tidy-interval", 30*time.Second, "interval at which to tidy-up OPC-UA subscriptions")
	fs.StringVar(&proxyListen, "proxy-listen", ":8080", "Centrifugo proxy listen address")
	fs.StringVar(&centrifugoClientConfig.Addr, "centrifugo-api-address", "", "Centrifugo API endpoint")
	fs.StringVar(&centrifugoNamespace, "centrifugo-namespace", "", "Centrifugo channel namespace for this instance")
	fs.StringVar(&centrifugoClientConfig.Key, "centrifugo-api-key", "", "Centrifugo API key")
	debug := fs.Bool("debug", false, "log debug information")
	fs.Usage = usageFor(fs, os.Stderr)

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		checkErr(log.With(logger, "during", "env file loading"), err, 1)
	}

	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarNoPrefix()); err != nil {
		checkErr(log.With(logger, "during", "flags parsing"), err, 2)
	}

	loglevel := level.AllowInfo()
	if *debug {
		loglevel = level.AllowDebug()
	}
	logger = level.NewFilter(logger, loglevel)

	var (
		opcClient  *opcua.Client
		opcMonitor *opcua.Monitor
	)
	func() {
		l := log.With(logger, "during", "OPC-UA client creation")
		ctx, cancel := context.WithTimeout(context.Background(), opcuaClientTimeout)
		defer cancel()
		sec, err := opcua.NewSecurity(&opcuaConfig, opcua.DefaultSecurityOptsProvider{})
		checkErr(l, err, 1)
		opcClient, err = opcua.NewClient(ctx, &opcuaConfig, opcua.DefaultClientExtDeps{}, sec)
		checkErr(l, err, 1)
		opcMonitor = opcua.NewMonitor(opcClient)
	}()

	proxy := proxy.NewProxy(opcMonitor, proxy.DefaultCentrifugoChannelParser{}, centrifugoNamespace)

	centrifugoClient := gocent.New(centrifugoClientConfig)

	var g run.Group

	{
		proxyLogger := log.With(logger, "component", "proxy")
		srv := http.Server{
			Addr:    proxyListen,
			Handler: proxy,
		}
		g.Add(func() error {
			defer level.Debug(proxyLogger).Log("msg", "shutting down")
			level.Debug(proxyLogger).Log("msg", "starting")
			level.Info(proxyLogger).Log("listen", proxyListen)
			return srv.ListenAndServe()
		}, func(err error) {
			ctx, cancel := context.WithTimeout(context.Background(), stopTimeout)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				level.Info(proxyLogger).Log("during", "Shutdown", "err", err)
			}
		})
	}

	{
		monitorLogger := log.With(logger, "component", "monitor")
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			defer level.Debug(monitorLogger).Log("msg", "stopping")
			level.Debug(monitorLogger).Log("msg", "starting")
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					c, d, err := opcMonitor.GetDataChange(ctx)
					if errors.Is(err, context.Canceled) {
						continue
					}
					if err != nil {
						level.Info(monitorLogger).Log("during", "GetDataChange", "err", err)
						continue
					}
					if _, err := centrifugoClient.Publish(ctx, c, d); err != nil {
						level.Info(monitorLogger).Log("during", "Publish", "err", err)
					}
				}
			}
		}, func(err error) {
			cancel()
			stopContext, stopCancel := context.WithTimeout(context.Background(), stopTimeout)
			defer stopCancel()
			errs := opcMonitor.Stop(stopContext)
			for _, err := range errs {
				level.Info(monitorLogger).Log("during", "stop", "err", err)
			}
		})
	}

	{
		tidyLogger := log.With(logger, "component", "tidy")
		ctx, cancel := context.WithCancel(context.Background())
		ticker := time.NewTicker(opcuaTidyInterval)
		g.Add(func() error {
			defer level.Debug(tidyLogger).Log("msg", "stopping")
			level.Debug(tidyLogger).Log("msg", "starting")
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-ticker.C:
					ints, err := Channels(ctx, centrifugoClient, centrifugoNamespace)
					if err != nil {
						level.Info(tidyLogger).Log("during", "ChannelIntervals", "err", err)
					}
					errs := opcMonitor.Purge(ctx, ints)
					for _, err := range errs {
						level.Info(tidyLogger).Log("during", "monitor purge", "err", err)
					}
				}
			}
		}, func(err error) {
			ticker.Stop()
			cancel()
		})
	}

	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	level.Info(logger).Log("exit", g.Run())
}
