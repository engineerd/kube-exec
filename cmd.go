package exec

import (
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"
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

	Cfg KubeConfig
	pod *v1.Pod

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Start starts the specified command but does not wait for it to complete.
func (cmd *Cmd) Start() error {

	pod, err := createPod(cmd.Cfg.Kubeconfig, cmd.Cfg.Namespace, cmd.Cfg.Name, cmd.Cfg.Image, []string{cmd.Path}, cmd.Args)
	if err != nil {
		return fmt.Errorf("cannot create pod: %v", err)
	}

	cmd.pod = pod
	fmt.Printf("created pod: %v\n", pod.Name)
	fmt.Printf("To wait the execution, use cmd.Wait() / cmd.Run(). To see the logs, use kubectl logs %v\n", pod.Name)

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
