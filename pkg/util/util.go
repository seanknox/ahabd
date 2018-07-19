package util

import (
	"os/exec"

	"github.com/juan-lee/ahabd/pkg/kubernetes"
)

func getHostName() (string, error) {
	out, err := exec.Command("hostname").CombinedOutput()
	if err != nil {
		return "", err
	}
	name := string(out)
	name = name[:len(name)-1]
	return name, nil
}

// GetNodeName returns the name of the node for this instance of the daemon
func GetNodeName() (string, error) {
	name, err := getHostName()
	if err != nil {
		return "", err
	}
	pod, err := kubernetes.GetPod(name, "kube-system")
	if err != nil {
		return "", err
	}

	return pod.Spec.NodeName, nil
}
