package main

import (
	"log"
	"os"

	kube "github.com/radu-matei/kube-exec"
)

func main() {

	cfg := kube.Config{
		Kubeconfig: os.Getenv("KUBECONFIG"),
		Image:      "ubuntu",
		Name:       "kube-example",
		Namespace:  "default",
	}

	// also sleeping for a couple of seconds
	// if the pod completes too fast, we don't have time to attach to it

	cmd := kube.Command(cfg, "/bin/sh", "-c", "sleep 2; echo Running from Kubernetes pod;")
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

}
