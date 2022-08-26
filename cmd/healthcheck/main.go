// Binary healthcheck.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func errExit(l log.Logger, err error) {
	level.Info(l).Log("err", err)
	os.Exit(1)
}

func main() {
	var (
		port uint
	)
	flag.UintVar(&port, "port", 0, "port the proxy listens on")
	flag.Parse()

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	var l log.Logger

	l = log.With(logger, "during", "flags validation")
	if port == 0 || port > uint(^uint16(0)) {
		errExit(l, fmt.Errorf("invalid port number: %d", port))
	}

	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	l = log.With(logger, "during", "health entrypoint request")
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
	if err != nil {
		errExit(l, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	l = log.With(logger, "from", "health response")
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			errExit(l, err)
		}
		errExit(l, errors.New(strings.TrimSpace(string(body))))
	}
}
