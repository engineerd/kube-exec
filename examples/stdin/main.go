package main

import (
	"io"
	"log"
	"os"

	kube "github.com/engineerd/kube-exec"
)

func main() {

	cfg := kube.Config{
		Kubeconfig: os.Getenv("KUBECONFIG"),
		Image:      "ubuntu",
		Name:       "kube-attach",
		Namespace:  "default",
	}

	cmd := kube.Command(cfg, "/bin/sh", "-c", "while true; read test; do echo You said: $test; sleep .5; done")

	w, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("cannot get pipe to stdin: %v", err)
	}

	// write in cmd.Stdin
	go func() {
		defer w.Close()
		_, err = io.Copy(w, os.Stdin)
		if err != nil {
			log.Fatalf("cannot copy from stdin: %v", err)
		}
	}()

	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

}
