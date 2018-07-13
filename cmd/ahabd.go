package cmd

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/juan-lee/ahabd/pkg/fixer"
	"github.com/juan-lee/ahabd/pkg/fixer/docker"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	version = "unreleased"

	// Command line flags
	period time.Duration
)

func Execute(ver string) {
	version = ver
	rootCmd := &cobra.Command{
		Use:   "ahabd",
		Short: "Docker Restart Daemon",
		Run:   root,
	}

	rootCmd.PersistentFlags().DurationVar(&period, "period", time.Minute*60,
		"restart check period")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func root(cmd *cobra.Command, args []string) {
	log.Infof("Docker Health Daemon: %s", version)
	log.Infof("Docker Health Check: every %v", period)

	go fixer.PeriodicFix(docker.New("ahabd"), period)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8081", nil))
}
