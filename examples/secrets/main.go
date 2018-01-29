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
		Name:       "secret",
		Namespace:  "default",

		Secrets: []kube.Secret{
			{
				EnvVarName: "SUPERPRIVATESECRET",
				SecretName: "k8s-secret",
				SecretKey:  "password",
			},
		},
	}

	cmd := kube.Command(cfg, "/bin/sh", "-c", "sleep 2; echo Your private secret: $SUPERPRIVATESECRET;")
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

}
