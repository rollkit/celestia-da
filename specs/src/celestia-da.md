# Celestia DA

## Abstract

This package implements the generic DA interface defined in [go-da] for Celestia.

## Details

The generic DA interface defines how DA implementations can submit, retrieve and validate blobs.

The Celestia implementation is a wrapper around [celestia-openrpc] which
connects to a local celestia node using an OpenRPC client.

A new client can be created by passing in the OpenRPC configuration. These include the following parameters:

| Parameter   | Description                                | JSON field  |
|-------------|--------------------------------------------|-------------|
| Auth Token  | Authentication Token for node connection   | auth_token  |
| Base URL    | URL for node connection                    | base_url    |
| Timeout     | Timeout for node connection                | timeout     |
| Fee         | Fee for submit blob transaction            | fee         |
| Gas         | Gas for the submit blob transaction        | gas_limit   |

## Assumptions

There should be a local celestia node, either full, bridge or light node running and accessible from the implementation.

The implementation uses the celestia node to interact with the celestia network to send, receive and validate blobs.

To be able to submit blobs, the following assumptions are made:

* The local celestia node is fully caught up with the network tip.
* The auth token has at least read/write permissions.
* The local celestia node has a account with a non zero balance to pay fee.

## Implementation

The implementation calls the corresponding Celestia [node api docs] methods.

### Get

Get retrieves blobs referred to by their ids.

The implementation calls [blob.Get] RPC method on the Celestia Node API.

### Commit

Commit returns the commitment to blobs.

The implementation calls `blob.CreateCommitments` which does not call any RPC method, so it's completely offline.

### Submit

Submit submits blobs and returns their ids and proofs.

The implementation calls [blob.Submit] RPC method with `DefaultSubmitOptions` on the Celestia Node API.

`DefaultSubmitOptions` uses default values for `Fee` and `GasLimit`.

### Validate

Validate validates blob ids and proofs and returns whether they are included.

The implementation calls [blob.Included] RPC method on the Celestia Node API.

## References
[1] [go-da]

[2] [celestia-openrpc]

[3] [node api docs]

[go-da]: https://github.com/rollkit/go-da
[celestia-openrpc]: https://github.com/rollkit/celestia-openrpc
[node api docs]: https://docs.celestia.org/api/v0.11.0-rc13/
[blob.Get]: https://docs.celestia.org/api/v0.11.0-rc13/#blob.Get
[blob.Submit]: https://docs.celestia.org/api/v0.11.0-rc13/#blob.Submit
[blob.Included]: https://docs.celestia.org/api/v0.11.0-rc13/#blob.Included
