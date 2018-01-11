kube-exec
=========

`kube-exec` is a library similar to [`os/exec`][1] that allows you to run commands in a Kubernetes pod, as if that command was executed locally.
> It is inspired from [`go-dexec`][2] by [ahmetb][3], which does the same thing, but for a Docker engine.

The interface of the package is similar to `os/exec`, and essentially this:

- creates a new pod in Kubernetes based on a user-specified image
- waits for the pod to be in [`Running`](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/) state
- attaches to the pod and allows you to stream data to the pod through `stdin`, and from the pod back to the program through `stdout` and `stderr`


How to use it
-------------

```go
cfg := kube.Config{
	Kubeconfig: os.Getenv("KUBECONFIG"),
	Image:      "ubuntu",
	Name:       "kube-example",
	Namespace:  "default",
}

// also sleeping for a couple of seconds
// if the pod completes too fast, we don't have time to catch the Running state event

cmd := kube.Command(cfg, "/bin/sh", "-c", "sleep 2; echo Running from Kubernetes pod;")
cmd.Stdout = os.Stdout

err := cmd.Run()
if err != nil {
	log.Fatalf("error: %v", err)
}
```


Here's a list of full examples you can find in this repo:

- [simple hello example](/examples/hello)
- [pass `stdin` to the pod](/examples/stdin)


[1]: https://golang.org/pkg/os/exec
[2]: https://github.com/ahmetb/go-dexec
[3]: https://twitter.com/ahmetb

[4]: /examples/main.go


Disclaimer
----------

This is not an official Microsoft project.