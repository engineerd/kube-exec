package main

import (
	"log"
	"os"

	kube "github.com/radu-matei/kube-exec"
)

var kubeconfig = os.Getenv("KUBECONFIG")

func main() {

	cfg := kube.Config{
		Kubeconfig: kubeconfig,
		Image:      "ubuntu",
		Name:       "kube-exec-attach-req-28",
		Namespace:  "default",
	}

	cmd := kube.Command(cfg, "/bin/sh", "-c", "while true; do echo hello; sleep .5; done")
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cannot start command: %v", err)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Wait()
	if err != nil {
		log.Fatalf("cannot wait command: %v", err)
	}
}
