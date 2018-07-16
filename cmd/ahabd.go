package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/juan-lee/ahabd/pkg/fixer"
	"github.com/juan-lee/ahabd/pkg/fixer/docker"
	"github.com/juan-lee/ahabd/pkg/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	// Command line flags
	period time.Duration
)

// Execute invokes the CLI.
func Execute() {
	rootCmd := &cobra.Command{
		Use:   "ahabd",
		Short: "Node Health Daemon",
		Run:   root,
	}

	rootCmd.PersistentFlags().DurationVar(&period, "period", time.Minute*60,
		"restart check period")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func root(cmd *cobra.Command, args []string) {
	log.Infof("Daemon Version: %s", version.Version)
	log.Infof("Health Check: every %v", period)

	g, ctx := errgroup.WithContext(newContext())
	g.Go(func() error {
		return fixer.PeriodicFix(ctx, docker.New("ahabd"), period)
	})
	g.Go(func() error {
		return fixer.PeriodicFix(ctx, fixer.NewKubeProxy("ahabd"), period)
	})
	g.Go(func() error {
		return serveMetrics(ctx)
	})

	if err := g.Wait(); err != nil {
		log.Warnf("Exiting with [%v]", err)
	}
}

func newContext() context.Context {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx
}

func serveMetrics(ctx context.Context) error {
	http.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{
		Handler: promhttp.Handler(),
		Addr:    ":8081",
	}

	go func() {
		log.Info(srv.ListenAndServe())
	}()

	<-ctx.Done()

	log.Infof("Shutting down metrics endpoint.")
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Warnf("Error shutting down metrics endpoint: %v", err)
		return err
	}

	return nil
}
