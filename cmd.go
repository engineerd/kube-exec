package exec

import (
	"fmt"
	"io"

	batch "k8s.io/api/batch/v1"
)

// KubeConfig contains all Kubernetes configuration
type KubeConfig struct {
	Kubeconfig string
	Namespace  string
	Name       string
	Image      string
}

// Cmd represents the command to execute inside the pod
type Cmd struct {
	Path string
	Args []string
	Env  []string
	Dir  string

	Cfg     KubeConfig
	kubeJob *batch.Job

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Run starts the specified command and waits for it to complete.
func (cmd *Cmd) Run() error {

	pod, err := createPod(cmd.Cfg.Kubeconfig, cmd.Cfg.Namespace, cmd.Cfg.Name, cmd.Cfg.Image, []string{cmd.Path}, cmd.Args)
	if err != nil {
		return fmt.Errorf("cannot create pod: %v", err)
	}

	fmt.Printf("created pod: %v", pod.Name)

	return nil
}

// Command returns the Cmd struct to execute the named program with
// the given arguments.
func Command(cfg KubeConfig, name string, arg ...string) *Cmd {
	return &Cmd{
		Cfg:  cfg,
		Path: name,
		Args: arg,
	}
}
