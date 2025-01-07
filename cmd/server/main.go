package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/victorzhuk/homyak/internal/app"
	"github.com/victorzhuk/homyak/internal/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	var (
		configPath string
		needHelp   bool
	)
	f := flag.NewFlagSet("app", 1)
	f.StringVar(&configPath, "config", "", "Path to configuration file")
	f.BoolVar(&needHelp, "help", false, "Print help")
	if err := f.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	if len(f.Args()) == 0 || needHelp {
		f.Usage()
		fmt.Println("\nCommands:")
		fmt.Println("  run\t\t\trun server\n  build\t\tbuild info\n  envs\t\tavailable envs")
		os.Exit(0)
	}

	cfg := app.MustParseConfig(configPath)
	app.SetupLogger(cfg)
	logger.Debug(fmt.Sprintf("current config: %#v", cfg))

	switch f.Args()[0] {
	case "run":
		break
	case "build":
		app.PrintBuildInfo(os.Stderr)
		os.Exit(0)
	case "envs":
		app.PrintEnvsUsage(os.Stderr)
		os.Exit(0)
	default:
		panic(fmt.Errorf("unknown command: %s", f.Args()[0]))
	}

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		<-sigs
		logger.Info("received shutdown signal")
		cancel()
	}()

	a := app.New(cfg)
	if err := a.Run(ctx); err != nil {
		logger.Fatal("app running failed", zap.Error(err))
	}

	os.Exit(0)
}
