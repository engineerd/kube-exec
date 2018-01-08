package main

import (
	"fmt"
	"os"

	kube "github.com/radu-matei/kube-exec"
)

var kubeconfig = os.Getenv("KUBECONFIG")

func main() {

	cfg := kube.KubeConfig{
		Kubeconfig: kubeconfig,
		Image:      "ubuntu",
		Name:       "kube-exec",
		Namespace:  "default",
	}

	cmd := kube.Command(cfg, "echo", "Hello, Universe!")
	err := cmd.Start()
	if err != nil {
		fmt.Printf("%v", err)
	}
}
