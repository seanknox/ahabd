package main

import (
	"math/rand"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/weaveworks/kured/pkg/delaytick"
)

var (
	version = "unreleased"

	// Command line flags
	period time.Duration
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ahabd",
		Short: "Docker Restart Daemon",
		Run:   root}

	rootCmd.PersistentFlags().DurationVar(&period, "period", time.Minute*60,
		"restart check period")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// newCommand creates a new Command with stdout/stderr wired to our standard logger
func newCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)

	cmd.Stdout = log.NewEntry(log.StandardLogger()).
		WithField("cmd", cmd.Args[0]).
		WithField("std", "out").
		WriterLevel(log.InfoLevel)

	cmd.Stderr = log.NewEntry(log.StandardLogger()).
		WithField("cmd", cmd.Args[0]).
		WithField("std", "err").
		WriterLevel(log.WarnLevel)

	return cmd
}

func restartRequired() bool {
	log.Infof("Checking if docker daemon needs restart")

	curlCmd := newCommand("curl", "--unix-socket", "/var/run/docker.sock", "http://localhost/info")
	if err := curlCmd.Run(); err != nil {
		log.Fatalf("Error invoking curl command: %v", err)
		return true
	}

	return false
}

func commandRestart(nodeID string) {
	log.Infof("Commanding docker daemon restart")

	// Relies on /var/run/dbus/system_bus_socket bind mount to talk to systemd
	restartCmd := newCommand("/bin/systemctl", "restart", "docker.service")
	if err := restartCmd.Run(); err != nil {
		log.Fatalf("Error invoking restart command: %v", err)
	}
}

func restartAsRequired(nodeID string) {
	source := rand.NewSource(time.Now().UnixNano())
	tick := delaytick.New(source, period)
	for _ = range tick {
		if restartRequired() {
			commandRestart(nodeID)
		}
	}
}

func root(cmd *cobra.Command, args []string) {
	log.Infof("Docker Restart Daemon: %s", version)

	nodeID := os.Getenv("AHABD_NODE_ID")
	if nodeID == "" {
		log.Fatal("AHABD_NODE_ID environment variable required")
	}

	log.Infof("Node ID: %s", nodeID)
	log.Infof("Docker Restart: every %v", period)

	go restartAsRequired(nodeID)

	time.Sleep(time.Hour)
}
