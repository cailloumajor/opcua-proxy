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

	go func() {
		s := <-sc
		signal.Stop(sc)
		log.Printf("received %v signal, cancelling main context", s)
		cancel()
	}()

	// TODO: remove
	<-ctx.Done()

}
