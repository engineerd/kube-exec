---
date: 2019-02-18
title: Contributing
---

## Contributing

We are delighted you want to contribute to this project! Keep in mind that any contribution to this project **MUST** adhere to the [Contributor Covenant Code of Conduct][coc].

Here's a short check list to help you contribute to this project:

- check the [issue queue][issues] and [PR queue][prs] to make sure you're not duplicating any developer's work.
- if there is an existing issue, please comment on it before starting to work on the implementation - if there isn't one, please create it.
- fork the project.
- follow the instructions below to make sure you have all required prerequisites to build the project.
- create a pull request with your changes.

Any contribution is extremely appreciated - documentation, bug fixes or features. Thank you!

## Prerequisites

- [the Go toolchain][go]
- [`dep`][dep]
- `make` (optional)

## Building from source

- `dep ensure`
- `make build` to build the library
- `make examples` to build all examples in `examples/`
- if running locally, you should provide an environment variable for the Kubernetes configuration file:
  - on Linux (including Windows Subsystem for Linux) and macOS: `export KUBECONFIG=<path-to-config>`
  - on Windows: `$env:KUBECONFIG="<path-to-config>"` 
- alternatively, you can individually `go run` the desired example locally, provided you pass a valid Kubernetes config file


[coc]: https://www.contributor-covenant.org/

[issues]: https://github.com/engineerd/kube-exec/issues
[prs]: https://github.com/engineerd/kube-exec/pulls

[go]: https://golang.org/doc/install
[dep]: https://github.com/golang/dep
