package main

import (
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/juan-lee/ahabd/pkg/docker"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/weaveworks/kured/pkg/delaytick"
)

const (
	testContainer = "ahabdtest"

	testImage    = "alpine"
	testImageTag = "latest"
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

func dockerRestartRequired() bool {
	log.Infof("Checking docker daemon health")

	if err := checkContainerLifecycle(); err != nil {
		return true
	}

	return false
}

func checkContainerLifecycle() error {
	log.Infof("Checking if we can run a container")

	docker := docker.New()

	err := docker.ImagePull(testImage, testImageTag)
	if err != nil {
		log.Warnf("Error pulling image: %v", err)
		return err
	}

	err = docker.ContainerDelete(testContainer)
	if err != nil {
		log.Warnf("Error deleting a container: %v", err)
		return err
	}

	err = docker.ContainerCreate(testContainer, testImage, testImageTag)
	if err != nil {
		log.Warnf("Error creating container: %v", err)
		return err
	}

	err = docker.ContainerStart(testContainer)
	if err != nil {
		log.Warnf("Error starting container: %v", err)
		return err
	}

	err = docker.ContainerWait(testContainer)
	if err != nil {
		log.Warnf("Error waiting for container: %v", err)
		return err
	}

	log.Infof("Confirmed we can run a container")

	return nil
}

func commandRestartDocker(nodeID string) error {
	log.Infof("Restarting docker daemon")

	// Relies on /var/run/dbus/system_bus_socket bind mount to talk to systemd
	restartCmd := newCommand("/bin/systemctl", "restart", "docker.service")
	if err := restartCmd.Run(); err != nil {
		log.Warnf("Error invoking docker restart command: %v", err)
		return err
	}

	return nil
}

func restartDockerAsRequired(nodeID string) {
	source := rand.NewSource(time.Now().UnixNano())
	tick := delaytick.New(source, period)
	for _ = range tick {
		if dockerRestartRequired() {
			commandRestartDocker(nodeID)
		}
	}
}

func root(cmd *cobra.Command, args []string) {
	log.Infof("Docker Restart Daemon: %s", version)

	nodeID := os.Getenv("AHABD_NODE_ID")
	if nodeID == "" {
		log.Warnf("AHABD_NODE_ID environment variable required")
	}

	log.Infof("Node ID: %s", nodeID)
	log.Infof("Docker Restart: every %v", period)

	restartDockerAsRequired(nodeID)
}
