package exec

import (
	"fmt"
	"io"
	"io/ioutil"

	v1 "k8s.io/api/core/v1"
)

// Config contains all Kubernetes configuration
type Config struct {
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

	Cfg Config
	pod *v1.Pod

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Command returns the Cmd struct to execute the named program with
// the given arguments.
func Command(cfg Config, name string, arg ...string) *Cmd {
	return &Cmd{
		Cfg:  cfg,
		Path: name,
		Args: arg,
	}
}

// Start starts the specified command but does not wait for it to complete.
func (cmd *Cmd) Start() error {
	if cmd.Stdin == nil {
		cmd.Stdin = ioutil.NopCloser(nil)
	}

	if cmd.Stdout == nil {
		cmd.Stdout = ioutil.Discard
	}

	if cmd.Stderr == nil {
		cmd.Stdout = ioutil.Discard
	}

	pod, err := createPod(cmd.Cfg.Kubeconfig, cmd.Cfg.Namespace, cmd.Cfg.Name, cmd.Cfg.Image, []string{cmd.Path}, cmd.Args)
	if err != nil {
		return fmt.Errorf("cannot create pod: %v", err)
	}

	cmd.pod = pod

	fmt.Printf("created pod: %v\n", pod.Name)
	fmt.Printf("To wait the execution, use cmd.Wait() / cmd.Run(). To see the logs, use kubectl logs %v\n", pod.Name)

	return nil
}

// Wait waits for the command to exit and waits for any copying to
// stdin or copying from stdout or stderr to complete.
//
// The command must have been started by Start.
func (cmd *Cmd) Wait() error {

	// wait for pod to be running
	fmt.Printf("waiting for pod to be running\n")
	watchPod(cmd.Cfg.Kubeconfig, cmd.pod)

	err := attach(cmd.Cfg.Kubeconfig, cmd.pod, cmd.Stdin, cmd.Stdout, cmd.Stderr)
	if err != nil {
		return fmt.Errorf("cannot attach: %v", err)
	}

	return nil
}

// StdinPipe returns a pipe that will be connected to the command's standard input
// when the command starts.
//
// Different than os/exec.StdinPipe, returned io.WriteCloser should be closed by user.
func (cmd *Cmd) StdinPipe() (io.WriteCloser, error) {

	pr, pw := io.Pipe()
	cmd.Stdin = pr
	return pw, nil
}
