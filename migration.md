# Migration

celestia-da after v0.13.2 release is deprecated and will no longer be
maintained.

Since go-da v0.5.0 now supports JSONRPC and celestia-node as of v0.13.0
implements the DA interface over JSONRPC, clients should migrate to using the
JSONRPC client instead.

To migrate to the JSONRPC client, the dial address must be configured to the
celestia-node >= v0.13.0 RPC server to connect to. An auth token with a blob
write permission is required for the connection. It can be obtained by using
the celestia command for e.g.:

```sh
export CELESTIA_NODE_AUTH_TOKEN=$(celestia light auth write)
```
Note that the auth token generated here should match the node type, network and
node store of the node being connected to.

```go
import (
    "os"
    "github.com/rollkit/go-da/proxy"
)

// url = "http://localhost:26658"
// auth_token = os.Getenv("CELESTIA_NODE_AUTH_TOKEN")

client, err := proxy.NewClient(ctx, url, auth_token)
...
client.MaxBlobSize()
// returns 1974272
```

See the [celestia-node release notes](https://github.com/celestiaorg/celestia-node/releases/tag/v0.13.0) for DA implementation details.
See the [go-da release notes](https://github.com/rollkit/go-da/releases/tag/v0.5.0) for details.
