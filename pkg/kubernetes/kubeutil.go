package exec

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Kubernetes represents configuration for cluster connection
type Kubernetes struct {
	KubeConfig  string
	KubeContext string
	Namespace   string

	Verbose bool

	Client *kubernetes.Clientset
}

// waitPod blocks until the pod is in finished state
//
// if the returned pod phase is succeeded, the caller should proceed to taking the logs
// if the phase is running, it can opt to attach to the pod standard output
func (k *Kubernetes) waitPod(name string, out io.Writer, label string) (*v1.PodPhase, error) {
	opts := meta.ListOptions{
		LabelSelector: label,
	}

	req, err := k.Client.CoreV1().Pods(k.Namespace).Watch(opts)
	if err != nil {
		return nil, err
	}

	res := req.ResultChan()

	timeout := time.After(10 * time.Minute)
	for {
		select {
		case e := <-res:
			if k.Verbose {
				obj, err := json.MarshalIndent(e.Object, "", " ")
				if err != nil {
					fmt.Fprintf(out, "cannot unmarshal object of type %v: %v", e.Type, err)
				}
				fmt.Fprintf(out, "Event: %s\n %s\n", e.Type, obj)
			}

			pod := e.Object.(*v1.Pod)

			switch e.Type {
			case "DELETED":
				return &pod.Status.Phase, fmt.Errorf("pod %v deleted unexpectedly", name)

			case "ADDED", "MODIFIED":
				switch pod.Status.Phase {
				case "Running", "Succeeded":
					req.Stop()
					return &pod.Status.Phase, nil

				case "Failed":
					req.Stop()
					return &pod.Status.Phase, fmt.Errorf("pod %v failed: %v", name, pod.Status.Reason)
				}
			}

		case <-timeout:
			req.Stop()
			return nil, fmt.Errorf("timeout waiting for pod %v to start", name)
		}
	}
}

// getKubeConfig returns the  Kubernetes REST configuration given a kubeconfig and context
func (k *Kubernetes) getKubeConfig() (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	rules.ExplicitPath = k.KubeConfig

	overrides := &clientcmd.ConfigOverrides{
		ClusterDefaults: clientcmd.ClusterDefaults,
		CurrentContext:  k.KubeContext,
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
}

// kubeClient returns an initialized Kubernetes client
func (k *Kubernetes) kubeClient() (*kubernetes.Clientset, error) {
	cfg, err := k.getKubeConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(cfg)
}
