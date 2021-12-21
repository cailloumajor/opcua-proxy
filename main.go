package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sc)

	// Termination loop
	go func() {
		for {
			select {
			case s := <-sc:
				log.Printf("received %v signal, cancelling main context", s)
				signal.Stop(sc)
				cancel()
			case <-ctx.Done():
				return
			}
		}
	}()
}
