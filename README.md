# celestia-da

celestia-da is an implementation of the [Generic DA interface](https://github.com/rollkit/go-da)
for modular blockchains. It extends celestia-node and runs a gRPC service,
which can be used by rollup clients to read and write blob data to a specific
namespace on celestia.

<!-- markdownlint-disable MD013 -->
[![build-and-test](https://github.com/rollkit/celestia-da/actions/workflows/ci_release.yml/badge.svg)](https://github.com/rollkit/celestia-da/actions/workflows/ci_release.yml)
[![golangci-lint](https://github.com/rollkit/celestia-da/actions/workflows/lint.yml/badge.svg)](https://github.com/rollkit/celestia-da/actions/workflows/lint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rollkit/celestia-da)](https://goreportcard.com/report/github.com/rollkit/celestia-da)
[![codecov](https://codecov.io/gh/rollkit/celestia-da/branch/main/graph/badge.svg?token=CWGA4RLDS9)](https://codecov.io/gh/rollkit/celestia-da)
[![GoDoc](https://godoc.org/github.com/rollkit/celestia-da?status.svg)](https://godoc.org/github.com/rollkit/celestia-da)
<!-- markdownlint-enable MD013 -->

## Minimum requirements

| Requirement | Notes          |
| ----------- |----------------|
| Go version  | 1.21 or higher |

## Installation

```sh
git clone https://github.com/rollkit/celestia-da.git
cd celestia-da
make build
sudo make install
```

## Usage

celestia-da is a wrapper around celestia-node, so see
[celestia node](https://github.com/celestiaorg/celestia-node) documentation for
details on configuring and running celestia-node.

celestia-da connects to celestia-node using JSON-RPC using the node rpc
endpoint. See [node rpc docs](https://node-rpc-docs.celestia.org/) for details.

celestia-da exposes a gRPC service that can be used with any gRPC client to
submit and retrieve blobs from a specific
namespace on the celestia network.

Note that celestia-da version may differ from the bundled celestia-node
version. Use the `celestia-da version` command to print the build information
including the bundled celestia-node version.

To start a celestia-da instance, use the preferred node type with `start`
command along with the gRPC specific flags as documented below.

## Example

Run celestia-da light mainnet node with a default DA interface server
accepting blobs on a randomly chosen namespace:

```sh
    celestia-da light start
        --core.ip <public ip>
        --da.grpc.namespace $(openssl rand -hex 10)
```

Note that the celestia-node RPC auth token is auto generated using the default
celestia-node store. If passed, the `da.grpc.token` flag
will override the default auth token.

## Flags

| Flag                         | Usage                                   | Default                     |
| ---------------------------- |-----------------------------------------|-----------------------------|
| `da.grpc.namespace`            | celestia namespace to use (hex encoded) | none; required              |
| `da.grpc.address`              | celestia-node RPC endpoint address      | `http://127.0.0.1:26658`      |
| `da.grpc.listen`               | gRPC service listen address             | `127.0.0.1:0`                 |
| `da.grpc.network`              | gRPC service listen network type        | `tcp`                         |
| `da.grpc.token`                | celestia-node RPC auth token            | `--node.store` auto generated |
| `da.grpc.gasprice`             | gas price for estimating fee (`utia/gas`) | -1 celestia-node default    |

See `celestia-da light/full/bridge start --help` for details.

### Tools

1. Install [golangci-lint](https://golangci-lint.run/usage/install/)
1. Install [markdownlint](https://github.com/DavidAnson/markdownlint)
1. Install [hadolint](https://github.com/hadolint/hadolint)
1. Install [yamllint](https://yamllint.readthedocs.io/en/stable/quickstart.html)

## Helpful commands

```sh
# Print celestia-da version build information, including bundled celestia-node version
celestia-da version

# Run unit tests
make test-unit

# Run all tests including integration tests
make test

# Run linters (requires golangci-lint, markdownlint, hadolint, and yamllint)
make lint
```

## Contributing

We welcome your contributions! Everyone is welcome to contribute, whether it's
in the form of code, documentation, bug reports, feature
requests, or anything else.

If you're looking for issues to work on, try looking at the
[good first issue list](https://github.com/rollkit/celestia-da/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).
Issues with this tag are suitable for a new external contributor and is a great
way to find something you can help with!

Please join our
[Community Discord](https://discord.com/invite/YsnTPcSfWQ)
to ask questions, discuss your ideas, and connect with other contributors.

## Code of Conduct

See our Code of Conduct [here](https://docs.celestia.org/community/coc).
