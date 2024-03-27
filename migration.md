### Migration

celestia-da after v0.13.2 release is deprecated and will no longer be
maintained.

Since go-da v0.5.0 now supports JSONRPC and celestia-node as of v0.13.0
implements the DA interface over JSONRPC, clients should migrate to using the
JSONRPC client instead.

A JSONRPC client requires the dial address and an auth token with a blob write
permission. It can be instantiated as follows:

```go
import (
    "github.com/rollkit/go-da/proxy/jsonrpc"
)

// url = "http://localhost:26658"
// auth_token = $(celestia light auth write)

client, err := jsonrpc.NewClient(ctx, url, auth_token)
client.DA.MaxBlobSize()
// returns 1974272
```

See the [go-da releasenotes](https://github.com/rollkit/go-da/releases/tag/v0.5.0) for details.
