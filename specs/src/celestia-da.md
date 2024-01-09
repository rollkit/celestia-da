# Celestia DA

## Abstract

This package implements the generic DA interface defined in [go-da] for Celestia.

## Details

The generic DA interface defines how DA implementations can submit, retrieve and validate blobs.

The Celestia implementation connects to a local [celestia-node] instance using a RPC client and allows using Celestia as the DA layer.

## Assumptions

There should be a local celestia node, either full, bridge or light node running and accessible from the implementation.

The implementation uses the celestia node to interact with the celestia network to send, receive and validate blobs.

To be able to submit blobs, the following assumptions are made:

* The local celestia node is fully caught up with the network tip.
* The auth token has at least read/write permissions.
* The local celestia node has an account with a non-zero balance to pay fees if it wants to send blobs or transactions. A balance is not required for retrieving blobs.

## Implementation

The implementation calls the corresponding Celestia [node api docs] methods.

### Get

Get retrieves blobs referred to by their ids.

The implementation calls [blob.Get] RPC method on the Celestia Node API.

### GetIDs

GetIDs returns the ids of all blobs at the given height.

The implementation calls [blob.GetAll] method on the Celestia Node API.

### Commit

Commit returns the commitment to blobs.

The implementation calls `blob.CreateCommitments` which does not call any RPC method, so it's completely offline.

### Submit

Submit submits blobs and returns their ids and proofs.

The implementation calls [blob.Submit] RPC method with `DefaultSubmitOptions` on the Celestia Node API if `gasPrice` is greater than or equal to zero.

`DefaultSubmitOptions` uses default values for `Fee` and `GasLimit`.

If `gasPrice` is less than zero, then it uses `app types` to `EstimateGas` based on the blob sizes and updates `GasLimit` and `Fee` on the `SubmitOptions` accordingly.

This way the client increase the `gasPrice`

### Validate

Validate validates blob ids and proofs and returns whether they are included.

The implementation calls [blob.Included] RPC method on the Celestia Node API.

## References

[1] [go-da]

[2] [celestia-node]

[3] [node api docs]

[go-da]: https://github.com/rollkit/go-da
[celestia-node]: https://github.com/celestiaorg/celestia-node
[node api docs]: https://node-rpc-docs.celestia.org/?version=v0.11.0
[blob.Get]: https://node-rpc-docs.celestia.org/?version=v0.11.0#blob.Get
[blob.GetAll]: https://node-rpc-docs.celestia.org/?version=v0.11.0#blob.GetAll
[blob.Submit]: https://node-rpc-docs.celestia.org/?version=v0.11.0#blob.Submit
[blob.Included]: https://node-rpc-docs.celestia.org/?version=v0.11.0#blob.Included
