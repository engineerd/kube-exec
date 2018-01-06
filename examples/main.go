package main

import (
	"os"

	exec "github.com/radu-matei/kube-exec"
)

var kubeconfig = os.Getenv("KUBECONFIG")

func main() {
	exec.PrintJobs(kubeconfig, "default")
}
