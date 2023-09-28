# Celestia DA

## Abstract

This package implements the generic DA interface defined in [go-da] (github.com/rollkit/go-da).

## Details

A new client can be created by passing in the OpenRPC configuration. These include the following parameters:

|  Parameter  |  Description                               |  JSON field
|  Auth Token |  Authentication Token for node connection  |  auth_token
|  Base URL   |  URL for node connection                   |  base_url
|  Timeout    |  Timeout for node connection               |  timeout
|  Fee        |  Fee for submit blob transaction           |  fee
|  Gas        |  Gas for the submit blob transaction       |  gas_limit

## Assumptions


## Implementation

### Get
Get retrieves Blobs referred to by their ids.

### Submit
Submit submits Blobs and returns their ids and proofs.

## References
[1] github.com/rollkit/go-da