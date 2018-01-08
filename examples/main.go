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
		Name:       "kube-testing123",
		Namespace:  "default",
	}

	cmd := kube.Command(cfg, "echo", "Hello, Universe!")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%v", err)
	}
}
