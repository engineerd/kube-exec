package main

import (
	"io"
	"log"
	"os"

	kube "github.com/radu-matei/kube-exec"
)

func main() {

	cfg := kube.Config{
		Kubeconfig: os.Getenv("KUBECONFIG"),
		Image:      "ubuntu",
		Name:       "kube-attach",
		Namespace:  "default",
	}

	cmd := kube.Command(cfg, "/bin/sh", "-c", "while true; read test; do echo hello; echo $test; sleep .5; done")
	//cmd := kube.Command(cfg, "/bin/sh", "-c", "while true; do echo hello; sleep .5; done")
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cannot start command: %v", err)
	}

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
	cmd.Stderr = os.Stderr

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("cannot wait command: %v", err)
	}

}
