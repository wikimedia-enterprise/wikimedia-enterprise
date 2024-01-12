# HTTP utilities

The repository contains frequently used `http` and `cognito authorization` helpers

# Auth

Function `Auth` creates authorization middleware.
It decodes access token and checks necessary claims, for example `ClientId`, uses `Provider` to get user and stores it in `Cache` for some passed period of time.

# JWK

`JWK` is a struct with basic AWS claims.
`JWK.RSA256()` generates and returns a public key.

# JWKProvider

`JWKProvider` is an interface which wraps `Fetch` and `Find` functions. `Fetch` gets keys from `/.well-known/jwks.json` and `Find` returns JWK with passed KID claim.
