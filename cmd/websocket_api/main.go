package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lillilli/logger"
	"github.com/lillilli/vconf"
	"github.com/pkg/errors"

	"github.com/lillilli/graphex/config"
	"github.com/lillilli/graphex/server"
	"github.com/lillilli/graphex/watcher"
)

var (
	configFile = flag.String("config", "", "set service config file")
)

func main() {
	flag.Parse()

	cfg := &config.Config{}

	if err := vconf.InitFromFile(*configFile, cfg); err != nil {
		fmt.Printf("unable to load config: %s\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.Log)
	log := logger.NewLogger("synchronizer")

	if err := startService(cfg); err != nil {
		log.Errorf("Start http server failed: %v", err)
	}
}

func startService(cfg *config.Config) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	watcher := watcher.New(cfg.WatchDir)
	if err := watcher.Start(ctx); err != nil {
		return errors.Wrap(err, "watch fs failed")
	}

	server := server.NewServer(cfg, watcher)
	if err := server.Start(); err != nil {
		return errors.Wrap(err, "listen ws failed")
	}

	<-signals
	close(signals)
	server.Stop()

	return nil
}
