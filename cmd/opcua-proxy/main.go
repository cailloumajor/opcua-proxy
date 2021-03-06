package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/cailloumajor/opcua-proxy/internal/centrifugo"
	"github.com/cailloumajor/opcua-proxy/internal/opcua"
	"github.com/cailloumajor/opcua-proxy/internal/proxy"
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

func errExit(l log.Logger, err error) {
	level.Info(l).Log("err", err)
	os.Exit(1)
}

type retryableInit struct {
	logger   log.Logger
	attempts uint
	maxDelay time.Duration
}

func (r *retryableInit) Do(f func(ctx context.Context) error) {
	rCtx, rCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer rCancel()

	err := retry.Do(
		func() error { return f(rCtx) },
		retry.Attempts(r.attempts),
		retry.Context(rCtx),
		retry.Delay(500*time.Millisecond),
		retry.LastErrorOnly(true),
		retry.MaxDelay(r.maxDelay),
		retry.OnRetry(func(n uint, err error) {
			level.Info(r.logger).Log("err", err, "retry", n)
		}),
	)
	if err != nil {
		errExit(r.logger, err)
	}
}

func main() {
	var (
		opcuaConfig            opcua.Config
		opcuaTidyInterval      time.Duration
		proxyListen            string
		centrifugoNamespace    string
		centrifugoClientConfig gocent.Config
		heartbeatInterval      time.Duration
		readNodesURL           string
	)
	fs := flag.NewFlagSet("opcua-proxy", flag.ExitOnError)
	fs.StringVar(&opcuaConfig.ServerURL, "opcua-server-url", "opc.tcp://127.0.0.1:4840", "OPC-UA server endpoint URL")
	fs.StringVar(&opcuaConfig.User, "opcua-user", "", "user name for OPC-UA authentication (optional)")
	fs.StringVar(&opcuaConfig.Password, "opcua-password", "", "password for OPC-UA authentication (optional)")
	fs.StringVar(&opcuaConfig.CertFile, "opcua-cert-file", "", "certificate file path for OPC-UA secure channel (optional)")
	fs.StringVar(&opcuaConfig.KeyFile, "opcua-key-file", "", "private key file path for OPC-UA secure channel (optional)")
	fs.DurationVar(&opcuaTidyInterval, "opcua-tidy-interval", 30*time.Second, "interval at which to tidy-up OPC-UA subscriptions")
	fs.StringVar(&proxyListen, "proxy-listen", ":8080", "Centrifugo proxy listen address")
	fs.StringVar(&centrifugoClientConfig.Addr, "centrifugo-api-address", "", "Centrifugo API endpoint")
	fs.StringVar(&centrifugoClientConfig.Key, "centrifugo-api-key", "", "Centrifugo API key")
	fs.StringVar(&centrifugoNamespace, "centrifugo-namespace", "", "Centrifugo channel namespace for this instance")
	fs.DurationVar(&heartbeatInterval, "heartbeat-interval", 5*time.Second, "Heartbeat interval")
	fs.StringVar(&readNodesURL, "read-nodes-url", "", "URL to query for nodes to read")
	debug := fs.Bool("debug", false, "log debug information")
	fs.Usage = usageFor(fs, os.Stderr)

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		errExit(log.With(logger, "during", "env file loading"), err)
	}

	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarNoPrefix()); err != nil {
		errExit(log.With(logger, "during", "flags parsing"), err)
	}

	{
		l := log.With(logger, "during", "flags validation")
		if err := ValidateCentrifugoAddress(centrifugoClientConfig.Addr); err != nil {
			errExit(log.With(l, "flag", "centrifugo-api-address"), err)
		}
		if centrifugoNamespace == "" {
			errExit(log.With(l, "flag", "centrifugo-namespace"), errors.New("missing namespace"))
		}
	}

	loglevel := level.AllowInfo()
	if *debug {
		loglevel = level.AllowDebug()
	}
	logger = level.NewFilter(logger, loglevel)

	var nodesToRead []opcua.NodesObject
	{
		r := &retryableInit{
			logger:   log.With(logger, "during", "getting nodes objects from URL"),
			attempts: 10,
			maxDelay: 2 * time.Second,
		}
		r.Do(func(ctx context.Context) error {
			var err error
			nodesToRead, err = opcua.NodesObjectsFromURL(ctx, readNodesURL)
			return err
		})
	}

	var opcClient *opcua.Client
	{
		sec, err := opcua.NewSecurity(&opcuaConfig, opcua.DefaultSecurityOptsProvider{})
		if err != nil {
			errExit(log.With(logger, "during", "OPC-UA security configuration"), err)
		}

		r := &retryableInit{
			logger:   log.With(logger, "during", "OPC-UA client creation"),
			attempts: 30,
			maxDelay: 10 * time.Second,
		}
		r.Do(func(ctx context.Context) error {
			tCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			var err error
			opcClient, err = opcua.NewClient(tCtx, &opcuaConfig, opcua.DefaultClientExtDeps{}, sec)
			return err
		})
	}

	opcMonitor := opcua.NewMonitor(opcClient, nodesToRead)

	centrifugoClient := gocent.New(centrifugoClientConfig)

	var g run.Group

	{
		proxyLogger := log.With(logger, "component", "proxy")
		proxy := proxy.NewProxy(
			proxyLogger,
			opcMonitor,
			proxy.DefaultCentrifugoChannelParser{},
			centrifugoClient,
			centrifugoNamespace,
		)
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
		heartbeatLogger := log.With(logger, "component", "heartbeat")
		ctx, cancel := context.WithCancel(context.Background())
		ticker := time.NewTicker(heartbeatInterval)
		g.Add(func() error {
			defer level.Debug(heartbeatLogger).Log("msg", "stopping")
			level.Debug(heartbeatLogger).Log("msg", "starting")
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-ticker.C:
					if err := centrifugo.PublishStatus(
						ctx,
						centrifugoNamespace,
						opcClient,
						centrifugoClient,
					); err != nil {
						level.Info(heartbeatLogger).Log("during", "status publishing", "err", err)
					}
				}
			}
		}, func(err error) {
			ticker.Stop()
			cancel()
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
					if !opcMonitor.HasSubscriptions() {
						continue
					}
					ints, err := Channels(ctx, centrifugoClient, centrifugoNamespace)
					if err != nil {
						level.Info(tidyLogger).Log("during", "Centrifugo channels query", "err", err)
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

	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	runErr := g.Run()

	var se run.SignalError
	if !errors.As(runErr, &se) {
		errExit(log.With(logger, "exit", "error"), runErr)
	}

	level.Info(logger).Log("exit", runErr)
}
