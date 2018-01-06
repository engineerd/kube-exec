package exec

import (
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetKubeClient is a convenience method for creating kubernetes config and client
// for a given kubeconfig
func GetKubeClient(kubeconfig string) (*kubernetes.Clientset, *restclient.Config, error) {
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

// PrintJobs returns all jobs in a given namespace
func PrintJobs(kubeconfig, namespace string) {
	clientset, _, err := GetKubeClient(kubeconfig)
	if err != nil {
		log.Fatalf("cannot get clientset: %v", err)
	}

	j, err := clientset.Batch().Jobs(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("cannot get jobs")
	}

	fmt.Printf("%v", j)
}
