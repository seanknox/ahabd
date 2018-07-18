package kubernetes

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetPod returns a corev1.Pod in a given namespace that matches a specified name
func GetPod(name string, namespace string) (*corev1.Pod, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return pod, nil
}

// GetPods returns a slice of corev1.Pods
func GetPods() ([]corev1.Pod, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	pl, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pl.Items, nil
}

// GetPodByPrefix will return the first pod where the name attribute matches the prefix supplied for a given node.
func GetPodByPrefix(prefix, nodeName string) (*corev1.Pod, error) {
	pods, err := GetPods()
	if err != nil {
		return nil, err
	}
	for _, p := range pods {
		if p.Spec.NodeName == nodeName {
			if strings.HasPrefix(p.Name, prefix) {
				return &p, nil
			}
		}
	}
	return nil, fmt.Errorf("A pod with prefix (%s) does not exist on node %s", prefix, nodeName)
}
