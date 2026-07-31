// Command mihomo-exporter exports Mihomo metrics for Prometheus.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap/zapcore"

	"github.com/etnperlong/mihomo-exporter/internal/collector"
	"github.com/etnperlong/mihomo-exporter/internal/config"
	"github.com/etnperlong/mihomo-exporter/internal/logging"
	"github.com/etnperlong/mihomo-exporter/internal/mihomo"
	"github.com/etnperlong/mihomo-exporter/internal/server"
)

const appName = "mihomo-exporter"

var (
	version          = "(devel)"
	errProberStopped = errors.New("prober stopped unexpectedly")
	errServerStopped = errors.New("HTTP server stopped unexpectedly")
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", appName, err)
		os.Exit(1)
	}
}

// run holds the real program so it stays testable and keeps main free of logic.
func run(ctx context.Context, args []string, stdout io.Writer) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("cancelled before start: %w", err)
	}

	cfg, err := config.Load(args)
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	if cfg.Version {
		if _, writeErr := fmt.Fprintln(stdout, buildVersion()); writeErr != nil {
			return fmt.Errorf("write version: %w", writeErr)
		}

		return nil
	}

	loggerOptions, err := newLoggingOptions(cfg.Log)
	if err != nil {
		return err
	}
	logger, err := logging.New(loggerOptions)
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}
	defer func() {
		// Syncing stderr on Linux can return EINVAL, so shutdown ignores this error.
		ignoreSyncError(logger.Sync())
	}()

	client, err := mihomo.New(cfg.Mihomo.APIURL, string(cfg.Mihomo.Secret), cfg.Mihomo.Timeout)
	if err != nil {
		return fmt.Errorf("create Mihomo client: %w", err)
	}
	store := collector.NewStore()
	prober, err := collector.NewProber(client, cfg.Probe, store, logger.Named("prober"))
	if err != nil {
		return fmt.Errorf("create prober: %w", err)
	}
	metricCollector, err := collector.NewCollector(store, cfg.Probe.Staleness)
	if err != nil {
		return fmt.Errorf("create collector: %w", err)
	}
	registry := prometheus.NewRegistry()
	if registerErr := registry.Register(metricCollector); registerErr != nil {
		return fmt.Errorf("register collector: %w", registerErr)
	}
	srv, err := server.New(cfg.Web, registry, logger.Named("server"))
	if err != nil {
		return fmt.Errorf("create HTTP server: %w", err)
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	proberDone := make(chan error, 1)
	go func() {
		proberDone <- prober.Run(runCtx)
	}()
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err := <-proberDone:
		if err != nil {
			return fmt.Errorf("run prober: %w", err)
		}

		return errProberStopped
	case err := <-serverDone:
		if err != nil {
			return err
		}

		return errServerStopped
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.Web.ShutdownTimeout)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shut down HTTP server: %w", err)
	}
	cancel()
	select {
	case err := <-proberDone:
		if err != nil {
			return fmt.Errorf("wait for prober: %w", err)
		}
	case <-shutdownCtx.Done():
		return fmt.Errorf("wait for prober: %w", shutdownCtx.Err())
	}
	if err := <-serverDone; err != nil {
		return err
	}

	return nil
}

func newLoggingOptions(logConfig config.LogConfig) (logging.Options, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logConfig.Level)); err != nil {
		return logging.Options{}, fmt.Errorf("parse log level %q: %w", logConfig.Level, err)
	}
	format, err := logging.ParseFormat(logConfig.Format)
	if err != nil {
		return logging.Options{}, fmt.Errorf("parse log format: %w", err)
	}
	emoji, err := logging.ParseToggle(logConfig.Emoji)
	if err != nil {
		return logging.Options{}, fmt.Errorf("parse log emoji: %w", err)
	}
	color, err := logging.ParseToggle(logConfig.Color)
	if err != nil {
		return logging.Options{}, fmt.Errorf("parse log color: %w", err)
	}

	return logging.Options{Level: level, Format: format, Emoji: emoji, Color: color, Output: os.Stderr}, nil
}

func buildVersion() string {
	if version != "" && version != "(devel)" {
		return version
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok && buildInfo.Main.Version != "" {
		return buildInfo.Main.Version
	}

	return "(devel)"
}

func ignoreSyncError(_ error) {}
