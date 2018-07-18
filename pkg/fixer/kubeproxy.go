package fixer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/juan-lee/ahabd/pkg/fixer/stats"
	"github.com/juan-lee/ahabd/pkg/kubernetes"
	"github.com/juan-lee/ahabd/pkg/util"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// kubeProxyFixer represents the fixer interface for repairing kubeproxy pods on a given cluster.
type kubeProxyFixer struct {
	stats stats.Stats
}

// Time holds a time.Time so we can override json.UnmarshalJSON
type Time struct {
	time.Time
}

// KubeProxyHealthz represents the json returned when hitting the healthz endpoint for kubeproxy
type KubeProxyHealthz struct {
	LastUpdated Time `json:"lastUpdated"`
	Current     Time `json:"currentTime"`
}

// NewKubeProxy returns a new kube proxy fixer responsible for monitoring and fixing kube proxy.
func NewKubeProxy(source string) Fixer {
	return &kubeProxyFixer{
		stats: stats.NewDefault(source, "kube_proxy"),
	}
}

// NeedsFixing returns true if the KubeProxy needs to be restarted
func (k *kubeProxyFixer) NeedsFixing(ctx context.Context) bool {
	log.Info("checking kube-proxy health")
	nodeName, err := util.GetNodeName()
	if err != nil {
		log.Printf("Error trying to fetch node name: %s", err)
		return false
	}
	kubeProxy, err := kubernetes.GetPodByPrefix("kube-proxy", nodeName)
	if err != nil {
		log.Printf("Error trying to get kube-proxy pod: %s", err)
		return false
	}

	return !isHealthy(kubeProxy)
}

// Fix performs the operation of fixing KubeProxy
func (k *kubeProxyFixer) Fix(ctx context.Context) error {
	log.Info("fixing kube-proxy -- This is a noop for now. We are observing only!")
	return nil
}

// Stats returns metrics about the operation
func (k *kubeProxyFixer) Stats() stats.Stats {
	return k.stats
}

func isHealthy(pod *corev1.Pod) bool {
	url := fmt.Sprintf("http://%s:%v/healthz", pod.Status.PodIP, 10256)
	resp, err := http.Get(url)
	if err != nil {
		return false
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	kubeProxyHealthz := KubeProxyHealthz{}
	err = json.Unmarshal(body, &kubeProxyHealthz)
	if err != nil {
		log.Printf("Error while trying to parse kube proxy healthz: %s", err)
		log.Printf("Response Body was: %s", body)
		return false
	}

	diff := kubeProxyHealthz.LastUpdated.Time.Sub(kubeProxyHealthz.Current.Time)
	if diff > 30*time.Second {
		log.Printf("Diff is > 30 seconds: %s", diff)
		log.Printf("Kube Proxy Health: %+v", kubeProxyHealthz)
		return false
	}
	return true
}

// UnmarshalJSON is a helper function to unmarshal non-standard time from JSON to struct
// 2018-07-17 20:15:40.752561974 +0000 UTC
func (t *Time) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	parsedTime, err := time.Parse("2006-01-02 15:04:05.999999 +0000 UTC", s)
	if err != nil {
		return err
	}
	t.Time = parsedTime
	return nil
}
