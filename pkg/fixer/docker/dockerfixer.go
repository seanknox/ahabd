package docker

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-systemd/dbus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/juan-lee/ahabd/pkg/fixer/stats"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	rebootSentinel = "/var/run/reboot-required"
)

type DockerFixer struct {
	stats     stats.Stats
	container containerRunner
	docker    serviceRestarter
	system    serviceRestarter
}

func New(source string) *DockerFixer {
	return &DockerFixer{
		stats:     stats.NewDefault(source, "docker"),
		container: &dockerContainerRunner{},
		docker:    &dockerRestarter{},
		system:    &systemRestarter{},
	}
}

func NewWithCounter(s stats.Stats) *DockerFixer {
	return &DockerFixer{
		stats:     s,
		container: &dockerContainerRunner{},
		docker:    &dockerRestarter{},
	}
}

func (df *DockerFixer) NeedsFixing(ctx context.Context) bool {
	log.Infof("checking docker daemon health")
	nf := make(chan bool)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	go func() {
		if err := df.container.Run(ctx); err != nil {
			log.Warnf("docker health check failed: %v", err)
			nf <- true
		}
		nf <- false
	}()

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return true
		}
		return false
	case result := <-nf:
		return result
	}
}

func (df *DockerFixer) Fix(ctx context.Context) error {
	err := df.docker.Restart(ctx)
	if err != nil {
		log.Warnf("restarting docker daemon failed: %v", err)
		return df.system.Restart(ctx)
	}
	return nil
}

func (df *DockerFixer) Stats() stats.Stats {
	return df.stats
}

type containerRunner interface {
	Run(ctx context.Context) error
}

type dockerContainerRunner struct{}

func (dcr *dockerContainerRunner) Run(ctx context.Context) error {
	log.Infof("running a container")
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	log.Infof("docker pull docker.io/alpine")
	reader, err := cli.ImagePull(ctx, "docker.io/library/alpine", types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	if b, err := ioutil.ReadAll(reader); err == nil {
		log.Infof(strings.TrimSpace(string(b)))
	}

	log.Infof("docker create alpine 'echo hello world'")
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "docker.io/library/alpine",
		Cmd:   []string{"echo", "hello world"},
		Tty:   true,
	}, nil, nil, "")
	if err != nil {
		return err
	}

	log.Infof("docker start %s", resp.ID)
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	log.Infof("docker wait %s", resp.ID)
	_, err = cli.ContainerWait(ctx, resp.ID)
	if err != nil {
		return err
	}

	log.Infof("docker logs %s", resp.ID)
	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return err
	}
	defer out.Close()

	if b, err := ioutil.ReadAll(out); err != nil || strings.TrimSpace(string(b)) != "hello world" {
		return errors.New(fmt.Sprintf("expected [hello world] got %v [%s]", err, string(b)))
	}

	log.Infof("docker rm %s", resp.ID)
	err = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}

	log.Infof("done running a container")
	return nil
}

type serviceRestarter interface {
	Restart(ctx context.Context) error
}

type dockerRestarter struct{}

func (dr *dockerRestarter) Restart(ctx context.Context) error {
	log.Warnf("restarting docker daemon")

	// Relies on /var/run/dbus/system_bus_socket bind mount to talk to systemd
	conn, err := dbus.New()
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Infof("systemctl restart docker.service")
	ch := make(chan string)
	_, err = conn.RestartUnit("docker.service", "replace", ch)
	if err != nil {
		return err
	}

	resp := <-ch
	if resp != "done" {
		return errors.New(fmt.Sprintf("couldn't restart docker - %s", resp))
	}

	log.Infof("done restarting docker daemon")
	return nil
}

type systemRestarter struct{}

func (sr *systemRestarter) Restart(ctx context.Context) error {
	log.Warnf("docker daemon requires reboot")

	if _, err := os.Stat(rebootSentinel); err == nil {
		log.Infof("node is already scheduled for reboot")
		return nil
	}

	err := ioutil.WriteFile(
		rebootSentinel,
		[]byte("*** System restart required (reason:docker) ***"),
		0644)
	if err != nil {
		log.Warnf("failed to write reboot sentinel %v", err)
		return err
	}

	return nil
}
