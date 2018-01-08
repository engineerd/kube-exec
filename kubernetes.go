package exec

import (
	"fmt"
	"log"

	"github.com/Azure/go-autorest/autorest/to"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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
					Name:    name,
					Image:   image,
					Command: command,
					Args:    args,
					SecurityContext: &v1.SecurityContext{
						Privileged: to.BoolPtr(false),
					},
					ImagePullPolicy: v1.PullPolicy(v1.PullIfNotPresent),
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
