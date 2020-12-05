package exec

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// Cmd represents the command to execute inside the pod
type Cmd struct {
	// Path is the path of the command to run.
	Path string

	// Args holds command line arguments, including the command as Args[0].
	// In typical use, both Path and Args are set by calling Command.
	Args []string

	// Env specifies the environment of the process.
	// TODO - decide going back of the []string of form "key=value".
	Env map[string]string

	// Dir specifies the working directory of the command.
	// If Dir is the empty string, Run runs the command in the
	// default directory where the container starts.
	Dir string

	// Cfg represents the Kubernetes configuration
	Cfg Config

	// Stdin, Stdout and Stderr specifie the process's standard input, output and error.
	// They are attached to the container after it is in running state.
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// The following fields are missing from the implementation.
	//
	// Since they are specific to running a command locally, they don't have
	// a correspondent in running commands in a container, or immediate utility
	// is not obvious. If needed, feel free to open an issue on the repo.
	//
	// // ExtraFiles specifies additional open files to be inherited by the
	// // new process. It does not include standard input, standard output, or
	// // standard error. If non-nil, entry i becomes file descriptor 3+i.
	// //
	// // ExtraFiles is not supported on Windows.
	// ExtraFiles []*os.File

	// // SysProcAttr holds optional, operating system-specific attributes.
	// // Run passes it to os.StartProcess as the os.ProcAttr's Sys field.
	// SysProcAttr *syscall.SysProcAttr

	// // Process is the underlying process, once started.
	// Process *os.Process

	// // ProcessState contains information about an exited process,
	// // available after a call to Wait or Run.
	// ProcessState *os.ProcessState
	// // contains filtered or unexported fields
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
	_, err := cmd.Cfg.CreatePod(cmd.Dir, []string{cmd.Path}, cmd.Args, cmd.Env)
	if err != nil {
		return fmt.Errorf("cannot create pod: %v", err)
	}
	return nil
}

// Wait waits for the command to exit and waits for any copying to
// stdin or copying from stdout or stderr to complete.
//
// The command must have been started by Start.
func (cmd *Cmd) Wait() error {
	if cmd.Stdin == nil {
		cmd.Stdin = ioutil.NopCloser(nil)
	}

	if cmd.Stdout == nil {
		cmd.Stdout = ioutil.Discard
	}

	if cmd.Stderr == nil {
		cmd.Stderr = ioutil.Discard
	}

	phase, err := cmd.Cfg.WaitPod(cmd.Cfg.Name, os.Stdout, "heritage=kube-exec")
	if err != nil {
		return fmt.Errorf("cannot wait for pod: %v", err)
	}

	// if pod is running, try to attach
	if *phase == v1.PodRunning {
		attachOptions := &v1.PodAttachOptions{
			Stdin:  cmd.Stdin != ioutil.NopCloser(nil),
			Stdout: cmd.Stdout != ioutil.Discard,
			Stderr: cmd.Stderr != ioutil.Discard,
			TTY:    false,
		}

		err := cmd.Cfg.Attach(attachOptions, cmd.Stdin, cmd.Stdout, cmd.Stderr)
		if err != nil {
			return fmt.Errorf("cannot attach: %v", err)
		}
	}

	// if pod has succeeded, try to get the logs
	if *phase == v1.PodSucceeded {

		c, err := containerToAttachTo("", cmd.Cfg.Pod)
		if err != nil {
			return fmt.Errorf("cannot get container for logs: %v", err)
		}
		logOptions := &v1.PodLogOptions{
			Container:  c.Name,
			Follow:     false,
			Timestamps: true,
		}

		client, err := cmd.Cfg.KubeClient()
		if err != nil {
			return fmt.Errorf("cannot create clientset: %v", err)
		}

		logStream, err := client.CoreV1().RESTClient().Get().
			Namespace(cmd.Cfg.Name).
			Name(cmd.Cfg.Name).
			Resource("pods").
			SubResource("log").
			VersionedParams(logOptions, scheme.ParameterCodec).Stream()
		if err != nil {
			return fmt.Errorf("pod finished, but cannot get logs: %v", err)
		}

		io.Copy(cmd.Stdout, logStream)
	}

	return nil
}

// Run starts the specified command and waits for it to complete.
func (cmd *Cmd) Run() error {
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("cannot start command: %v", err)
	}

	return cmd.Wait()
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
