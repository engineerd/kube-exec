package main

import (
	"log"
	"os"

	kube "github.com/engineerd/kube-exec"
)

func main() {
	cfg := kube.Config{
		KubeConfig: os.Getenv("KUBECONFIG"),
		Image:      "ubuntu",
		Name:       "secret",
		Namespace:  "default",
	}

	cmd := kube.Command(cfg, "/bin/sh", "-c", "sleep 2; echo Your private secret: $SUPERPRIVATESECRET;")
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

}
