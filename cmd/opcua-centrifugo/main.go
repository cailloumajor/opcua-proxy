package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cailloumajor/opcua-centrifugo/internal/config"
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

	ci := &config.DefaultInitializer{}
	cfg, err := config.Init(ci)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: remove lines below
	fmt.Println(cfg)
	<-ctx.Done()
}
