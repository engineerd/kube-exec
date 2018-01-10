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

	cmd := kube.Command(cfg, "/bin/sh", "-c", "sleep 1; echo Running from Kubernetes pod;")
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cannot start command: %v", err)
	}

	cmd.Stdout = os.Stdout
	err = cmd.Wait()
	if err != nil {
		log.Fatalf("cannot wait command: %v", err)
	}
}
