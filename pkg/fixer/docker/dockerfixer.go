package docker

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/coreos/go-systemd/dbus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/juan-lee/ahabd/pkg/fixer/stats"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	testContainer = "ahabdtest"

	testImage    = "alpine"
	testImageTag = "latest"
)

var (
	stdout = log.NewEntry(log.StandardLogger()).
		WithField("std", "out").
		WriterLevel(log.InfoLevel)
)

type DockerFixer struct {
	stats     stats.Stats
	container containerRunner
	docker    serviceRestarter
}

func New(source string) *DockerFixer {
	return &DockerFixer{
		stats:     stats.NewDefault(source, "docker"),
		container: &dockerContainerRunner{},
		docker:    &dockerRestarter{},
	}
}

func NewWithCounter(s stats.Stats) *DockerFixer {
	return &DockerFixer{
		stats:     s,
		container: &dockerContainerRunner{},
		docker:    &dockerRestarter{},
	}
}

func (df *DockerFixer) NeedsFixing() bool {
	log.Infof("Checking docker daemon health")

	if err := df.container.Run(); err != nil {
		return true
	}

	return false
}

func (df *DockerFixer) Fix() error {
	return df.docker.Restart()
}

func (df *DockerFixer) Stats() stats.Stats {
	return df.stats
}

type containerRunner interface {
	Run() error
}

type dockerContainerRunner struct{}

func (dcr *dockerContainerRunner) Run() error {
	log.Infof("Checking if we can run a container")

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	reader, err := cli.ImagePull(ctx, "docker.io/library/alpine", types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "alpine",
		Cmd:   []string{"echo", "hello world"},
		Tty:   true,
	}, nil, nil, "")
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	_, err = cli.ContainerWait(ctx, resp.ID)
	if err != nil {
		return err
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return err
	}
	defer out.Close()

	if b, err := ioutil.ReadAll(out); err != nil || strings.TrimSpace(string(b)) != "hello world" {
		return errors.New(fmt.Sprintf("expected [hello world] got %v [%s]", err, string(b)))
	}

	err = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}

	log.Infof("Confirmed we can run a container")
	return nil
}

type serviceRestarter interface {
	Restart() error
}

type dockerRestarter struct{}

func (dr *dockerRestarter) Restart() error {
	log.Infof("Restarting docker daemon")

	// Relies on /var/run/dbus/system_bus_socket bind mount to talk to systemd
	conn, err := dbus.New()
	if err != nil {
		return err
	}
	defer conn.Close()

	ch := make(chan string)
	_, err = conn.RestartUnit("docker.service", "replace", ch)
	if err != nil {
		return err
	}

	resp := <-ch
	if resp != "done" {
		return errors.New(fmt.Sprintf("couldn't restart docker - %s", resp))
	}

	return nil
}
