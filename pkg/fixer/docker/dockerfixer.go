package docker

import (
	"os/exec"

	"github.com/juan-lee/ahabd/pkg/docker"
	"github.com/juan-lee/ahabd/pkg/fixer/stats"
	log "github.com/sirupsen/logrus"
)

const (
	testContainer = "ahabdtest"

	testImage    = "alpine"
	testImageTag = "latest"
)

type DockerFixer struct {
	stats stats.Stats
}

func New(source string) *DockerFixer {
	return &DockerFixer{stats: stats.NewDefault(source, "docker")}
}

func NewWithCounter(s stats.Stats) *DockerFixer {
	return &DockerFixer{stats: s}
}

func (f DockerFixer) NeedsFixing() bool {
	return dockerRestartRequired()
}

func (f DockerFixer) Fix() error {
	return f.restartDocker()
}

func (f DockerFixer) Stats() stats.Stats {
	return f.stats
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

	if err := dockerHealthCheck(); err != nil {
		return true
	}

	return false
}

func dockerHealthCheck() error {
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

func (f DockerFixer) restartDocker() error {
	log.Infof("Restarting docker daemon")

	// Relies on /var/run/dbus/system_bus_socket bind mount to talk to systemd
	restartCmd := newCommand("/bin/systemctl", "restart", "docker.service")
	if err := restartCmd.Run(); err != nil {
		log.Warnf("Error invoking docker restart command: %v", err)
		return err
	}

	return nil
}
