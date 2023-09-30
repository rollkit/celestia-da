# Celestia DA

## Abstract

This package implements the generic DA interface defined in [go-da] (github.com/rollkit/go-da) for Celestia.

## Details

The generic DA inteface defines how DA implementations can submit, retrieve and validate blobs.

The Celestia implementation is a wrapper around [celestia-openrpc] (github.com/rollkit/celestia-openrpc) which
connects to a remote celestia node using a OpenRPC client.

A new client can be created by passing in the OpenRPC configuration. These include the following parameters:

|  Parameter  |  Description                               |  JSON field
|  Auth Token |  Authentication Token for node connection  |  auth_token
|  Base URL   |  URL for node connection                   |  base_url
|  Timeout    |  Timeout for node connection               |  timeout
|  Fee        |  Fee for submit blob transaction           |  fee
|  Gas        |  Gas for the submit blob transaction       |  gas_limit

## Assumptions

There should be a remote celestia node, either full, bridge or light node running and accessible from the client.

The client trusts the remote celestia node to connect to the celestia network and send, receive and validate messages.

To be able to submit blobs, the following assumptions are made:

    * The remote celestia node is fully caught up with the network to.
    * The auth token has atleast read/write permissions.
    * The remote celestia node has a account with a non zero balance to pay fee.

## Implementation

The implementation calls the corresponding [Celestia Node RPC API](https://docs.celestia.org/api/v0.11.0-rc13) methods.

### Get

Get retrieves blobs referred to by their ids.

The implementation calls [blob.Get](https://docs.celestia.org/api/v0.11.0-rc13/#blob.Get) RPC method on the Celestia Node API.

### Commit

Commit returns the commitment to blobs.

The implementation calls `blob.CreateCommitments` which does not call any RPC method, so it's completely offline.

### Submit

Submit submits blobs and returns their ids and proofs.

The implementation calls [blob.Submit](https://docs.celestia.org/api/v0.11.0-rc13/#blob.Submit) RPC method on the Celestia Node API.

### Validate

Validate validates blob ids and proofs and returns whether they are included.

The implementation calls [blob.Included](https://docs.celestia.org/api/v0.11.0-rc13/#blob.Included) RPC method on the Celestia Node API.

## References
[1] github.com/rollkit/go-da
[2] github.com/rollkit/celestia-openrpc
[3] https://docs.celestia.org/api/v0.11.0-rc13/
