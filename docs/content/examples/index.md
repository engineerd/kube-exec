---
date: 2019-02-18
title: Examples
---


## The simplest example

The following example creates a new pod based on the official Ubuntu image, then simply prints a message.

> Note: Because of a [known temporary limitation][log-issue], if the execution time of the command started inside the pod is too small, the logs will not be visible when executing, but have to be gathered using `kubectl`. Because of this reason, in this example there is a `sleep` command as well. This does **not** affect functionality, the command can be executed without the additional sleep statement, and this will be improved in the future.

[embedmd]:# (../../examples/hello/main.go go)
```go
package main

import (
	"log"
	"os"

	kube "github.com/engineerd/kube-exec"
)

func main() {

	cfg := kube.Config{
		Kubeconfig: os.Getenv("KUBECONFIG"),
		Image:      "ubuntu",
		Name:       "kube-example",
		Namespace:  "default",
	}

	// also sleeping for a couple of seconds
	// if the pod completes too fast, we don't have time to attach to it

	cmd := kube.Command(cfg, "/bin/sh", "-c", "sleep 2; echo Running from Kubernetes pod;")
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

}
```

## Passing `stdin` to the pod

Because this package follows the implementation of `os/exec`, you can also pass a pipe to the started pod and use it as standard input, as well as standard output, and standard error.

[embedmd]:# (../../examples/stdin/main.go go /func main/ $)
```go
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
```

## Passing secrets to the pod

If you need secret values to use inside your pod, Kubernetes secrets are available as environment variables.

[embedmd]:# (../../examples/secrets/main.go go /func main/ $)
```go
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
```


[log-issue]: https://github.com/engineerd/kube-exec/issues/6
