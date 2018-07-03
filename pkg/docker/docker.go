package docker

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
)

type DockerClient struct {
	client *http.Client
}

func New() *DockerClient {
	fd := func(proto, addr string) (conn net.Conn, err error) {
		return net.Dial("unix", "/var/run/docker.sock")
	}

	tr := &http.Transport{
		Dial: fd,
	}

	return &DockerClient{client: &http.Client{Transport: tr}}
}

func succeeded(statusCode int) bool {
	return statusCode >= 200 && statusCode <= 299
}

func (d *DockerClient) ImagePull(image string, tag string) error {
	resp, err := d.client.Post(
		fmt.Sprintf("http://localhost/images/create?fromImage=%s&tag=%s", image, tag),
		"application/json",
		nil)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	if !succeeded(resp.StatusCode) {
		return errors.New(fmt.Sprintf("image pull failed: %d - %s", resp.StatusCode, b))
	}

	return nil
}

func (d *DockerClient) ContainerCreate(name, image, tag string) error {
	resp, err := d.client.Post(
		fmt.Sprintf("http://localhost/containers/create?name=%s", name),
		"application/json",
		bytes.NewBuffer([]byte(fmt.Sprintf(`{"Image": "%s:%s", "Cmd": ["echo", "I'm", "alive"]}`, image, tag))))
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	if !succeeded(resp.StatusCode) {
		return errors.New(fmt.Sprintf("create container failed: %d - %s", resp.StatusCode, string(b)))
	}

	return nil
}

func (d *DockerClient) ContainerStart(name string) error {
	resp, err := d.client.Post(fmt.Sprintf("http://localhost/containers/%s/start", name), "application/json", nil)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	if !succeeded(resp.StatusCode) {
		return errors.New(fmt.Sprintf("start container failed: %d - %s", resp.StatusCode, b))
	}

	return nil
}

func (d *DockerClient) ContainerWait(name string) error {
	resp, err := d.client.Post(fmt.Sprintf("http://localhost/containers/%s/wait", name), "application/json", nil)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	if !succeeded(resp.StatusCode) {
		return errors.New(fmt.Sprintf("start container failed: %d - %s", resp.StatusCode, b))
	}

	return nil
}

func (d *DockerClient) ContainerDelete(name string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost/containers/%s", name), nil)
	if err != nil {
		return err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	if !succeeded(resp.StatusCode) {
		return errors.New(fmt.Sprintf("start container failed: %d - %s", resp.StatusCode, b))
	}

	return nil
}
