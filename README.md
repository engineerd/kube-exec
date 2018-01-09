kube-exec
=========

`kube-exec` is a library similar to [`os/exec`][1] that allows you to run commands in a Kubernetes pod, as if that command was executed locally.
> It is inspired from [`go-dexec`][2] by [ahmetb][3], which does the same thing, but for a Docker engine.

The interface of the package is similar to `os/exec`, and essentially this:

- creates a new pod in Kubernetes based on a user-specified image
- waits for the pod to be in [`Running`](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/) state
- attaches to the pod and allows you to stream data to the pod through `stdin`, and from the pod back to the program through `stdout` and `stderr`

You can [find an example in here][4] 

[1]: https://golang.org/pkg/os/exec
[2]: https://github.com/ahmetb/go-dexec
[3]: https://twitter.com/ahmetb

[4]: /examples/main.go