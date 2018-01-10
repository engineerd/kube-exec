package exec

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"time"

	"github.com/Azure/go-autorest/autorest/to"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

// getKubeClient is a convenience method for creating kubernetes config and client
// for a given kubeconfig
func getKubeClient(kubeconfig string) (*kubernetes.Clientset, *restclient.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get kubernetes config from kubeconfig '%s': %v", kubeconfig, err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get kubernetes client: %s", err)
	}
	return clientset, config, nil
}

// getPod returns a pod, given a namespace and pod name
func getPod(kubeconfig, namespace, name string) (*v1.Pod, error) {
	clientset, _, err := getKubeClient(kubeconfig)
	if err != nil {
		log.Fatalf("cannot get clientset: %v", err)
	}

	podsClient := clientset.CoreV1().Pods(namespace)

	return podsClient.Get(name, metav1.GetOptions{})
}

// createPod creates a new pod within a namespaces, with specified image and command to run
func createPod(kubeconfig, namespace, name, image string, command, args []string) (*v1.Pod, error) {
	clientset, _, err := getKubeClient(kubeconfig)
	if err != nil {
		log.Fatalf("cannot get clientset: %v", err)
	}

	podsClient := clientset.CoreV1().Pods(namespace)
	return podsClient.Create(&v1.Pod{

		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					TTY:   false,
					Stdin: true,

					Name:    name,
					Image:   image,
					Command: command,
					Args:    args,
					SecurityContext: &v1.SecurityContext{
						Privileged: to.BoolPtr(false),
					},
					ImagePullPolicy: v1.PullPolicy(v1.PullAlways),
					Env:             []v1.EnvVar{},
					VolumeMounts:    []v1.VolumeMount{},
				},
			},
			RestartPolicy:    v1.RestartPolicyOnFailure,
			Volumes:          []v1.Volume{},
			ImagePullSecrets: []v1.LocalObjectReference{},
		},
	})
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

// attach attaches to a given pod, outputting to stdout and stderr
func attach(kubeconfig string, pod *v1.Pod, attachOptions *v1.PodAttachOptions, stdin io.Reader, stdout, stderr io.Writer) error {
	clientset, config, err := getKubeClient(kubeconfig)
	if err != nil {
		log.Fatalf("cannot get clientset: %v", err)
	}

	container, err := containerToAttachTo("", pod)
	if err != nil {
		return fmt.Errorf("cannot get container to attach to: %v", err)
	}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("attach")

	attachOptions.Container = container.Name
	req.VersionedParams(attachOptions, scheme.ParameterCodec)

	streamOptions := getStreamOptions(attachOptions, stdin, stdout, stderr)

	err = startStream("POST", req.URL(), config, streamOptions)
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

// watchPod waits until the created pod is in running state
func watchPod(kubeconfig string, pod *v1.Pod) {
	clientset, _, err := getKubeClient(kubeconfig)
	if err != nil {
		log.Fatalf("cannot get clientset: %v", err)
	}

	stop := make(chan struct{})

	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", pod.Namespace, fields.Everything())
	_, controller := cache.NewInformer(watchlist, &v1.Pod{}, time.Second*1, cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(o, n interface{}) {
			newPod := n.(*v1.Pod)

			// not the pod we created
			if newPod.Name != pod.Name {
				return
			}

			// if the pod is running, stop watching and continue with the cmd execution
			if newPod.Status.Phase == v1.PodRunning {
				close(stop)
				return
			}
		},
	})

	controller.Run(stop)
}

func getStreamOptions(attachOptions *v1.PodAttachOptions, stdin io.Reader, stdout, stderr io.Writer) remotecommand.StreamOptions {
	fmt.Printf("attach options: %v", attachOptions)
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

	fmt.Printf("stream options: %v", streamOptions)
	return streamOptions
}
