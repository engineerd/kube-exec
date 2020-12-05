package exec

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"time"

	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

// Config represents Kubernetes configuration
type Config struct {
	KubeConfig  string
	KubeContext string
	Namespace   string
	Verbose     bool

	Name      string
	Image     string
	Labels    map[string]string
	Client    *kubernetes.Clientset
	Pod       *v1.Pod
	Container string
}

// CreatePod creates a new pod given basic configuration
//
// If anything else needs to be configured, use CreateWithPod
func (k *Config) CreatePod(workingDir string, command, args []string, env map[string]string) (*v1.Pod, error) {
	clientset, err := k.KubeClient()
	if err != nil {
		return nil, fmt.Errorf("cannot create clientset: %v", err)
	}

	c := clientset.CoreV1().Pods(k.Namespace)

	k.Pod = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   k.Name,
			Labels: appendDefaultLabel(k.Labels),
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyOnFailure,
			Containers: []v1.Container{
				{
					Stdin:      true,
					Name:       k.Name,
					Image:      k.Image,
					Command:    command,
					Args:       args,
					WorkingDir: workingDir,
					Env:        k8sEnv(env),
				},
			},
		},
	}

	return c.Create(k.Pod)
}

// CreateWithPod creates a new pod given a full pod definition
//
// If used, ALL values for a pod need to be provided, including
// name, image and labels, even if they are part of Config
func (k *Config) CreateWithPod(pod *v1.Pod) (*v1.Pod, error) {
	var err error

	k.Client, err = k.KubeClient()
	if err != nil {
		return nil, fmt.Errorf("cannot create clientset: %v", err)
	}

	c := k.Client.CoreV1().Pods(k.Namespace)
	pod.ObjectMeta.Labels = appendDefaultLabel(pod.ObjectMeta.Labels)
	k.Pod = pod
	return c.Create(pod)
}

func k8sEnv(env map[string]string) []v1.EnvVar {
	envVar := []v1.EnvVar{}
	for name, value := range env {
		envVar = append(envVar, v1.EnvVar{
			Name:  name,
			Value: value,
		})
	}

	return envVar
}

func appendDefaultLabel(l map[string]string) map[string]string {
	if len(l) == 0 {
		l = make(map[string]string)
	}

	l["heritage"] = "kube-exec"
	return l
}

// WaitPod blocks until the pod is in finished state
//
// if the returned pod phase is succeeded, the caller should proceed to taking the logs
// if the phase is running, it can opt to attach to the pod standard output
func (k *Config) WaitPod(name string, out io.Writer, label string) (*v1.PodPhase, error) {
	opts := metav1.ListOptions{
		LabelSelector: label,
	}

	client, err := k.KubeClient()
	if err != nil {
		return nil, fmt.Errorf("cannot create clientset: %v", err)
	}

	req, err := client.CoreV1().Pods(k.Namespace).Watch(opts)
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

// GetKubeConfig returns the  Kubernetes REST configuration given a kubeconfig and context
func (k *Config) GetKubeConfig() (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	rules.ExplicitPath = k.KubeConfig

	overrides := &clientcmd.ConfigOverrides{
		ClusterDefaults: clientcmd.ClusterDefaults,
		CurrentContext:  k.KubeContext,
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
}

// KubeClient returns an initialized Kubernetes client
func (k *Config) KubeClient() (*kubernetes.Clientset, error) {
	cfg, err := k.GetKubeConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(cfg)
}

// Attach attaches to a given pod, outputting to stdout and stderr
func (k *Config) Attach(attachOptions *v1.PodAttachOptions, stdin io.Reader, stdout, stderr io.Writer) error {
	clientset, err := k.KubeClient()
	if err != nil {
		log.Fatalf("cannot get clientset: %v", err)
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", k.KubeConfig)
	if err != nil {
		return fmt.Errorf("cannot get Kubernetes REST config: %v", err)
	}

	container, err := containerToAttachTo("", k.Pod)
	if err != nil {
		return fmt.Errorf("cannot get container to attach to: %v", err)
	}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(k.Pod.Name).
		Namespace(k.Pod.Namespace).
		SubResource("attach")

	attachOptions.Container = container.Name
	req.VersionedParams(attachOptions, scheme.ParameterCodec)

	streamOptions := getStreamOptions(attachOptions, stdin, stdout, stderr)

	err = startStream("POST", req.URL(), cfg, streamOptions)
	if err != nil {
		return fmt.Errorf("error executing: %v", err)
	}

	return nil
}

func startStream(method string, url *url.URL, config *restclient.Config, streamOptions remotecommand.StreamOptions) error {
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return err
	}

	return exec.Stream(streamOptions)
}

// containerToAttach returns a reference to the container to attach to, given
// by name or the first container if name is empty.
func containerToAttachTo(container string, pod *v1.Pod) (*v1.Container, error) {
	if len(container) > 0 {
		for i := range pod.Spec.Containers {
			if pod.Spec.Containers[i].Name == container {
				return &pod.Spec.Containers[i], nil
			}
		}
		for i := range pod.Spec.InitContainers {
			if pod.Spec.InitContainers[i].Name == container {
				return &pod.Spec.InitContainers[i], nil
			}
		}
		return nil, fmt.Errorf("container not found (%s)", container)
	}
	return &pod.Spec.Containers[0], nil
}

func getStreamOptions(attachOptions *v1.PodAttachOptions, stdin io.Reader, stdout, stderr io.Writer) remotecommand.StreamOptions {
	var streamOptions remotecommand.StreamOptions
	if attachOptions.Stdin {
		streamOptions.Stdin = stdin
	}

	if attachOptions.Stdout {
		streamOptions.Stdout = stdout
	}

	if attachOptions.Stderr {
		streamOptions.Stderr = stderr
	}

	return streamOptions
}
