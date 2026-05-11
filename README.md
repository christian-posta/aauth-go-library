# aauth-go-library

Go implementation of the AAuth protocol — Authenticated Authorization for autonomous agents.

This library provides the resource-side verification primitives and the agent-side client SDK for building AAuth-compliant services and agents. It is the Go counterpart to the Python aauth library.

## Packages

| Package | Description |
|---|---|
| `pkg/aauth` | Core resource-side API: `Verify`, `Challenge`, `MintResourceToken` |
| `pkg/aauth/agent` | Agent SDK: request signing, token exchange, deferred polling |
| `pkg/aauth/headers` | AAuth HTTP header parsing and serialization |
| `pkg/aauth/http` | 202 Deferred response helpers |
| `pkg/aauth/identifiers` | AAuth server/agent identifier validation |
| `pkg/aauth/keys` | JWKS fetcher with caching, JWK conversion |
| `pkg/aauth/metadata` | Well-known metadata document types and fetcher |
| `pkg/aauth/transport` | `http.RoundTripper` that signs outbound requests |
| `pkg/httpsig` | RFC 9421 HTTP Message Signatures (sign + verify) |
| `pkg/sigkey` | Signature-Key header parsing |

## Installation

```bash
go get github.com/christian-posta/aauth-go-library
```

## Usage

### Resource server: verifying an inbound request

```go
import (
    "github.com/christian-posta/aauth-go-library/pkg/aauth"
)

opts := aauth.VerifyOptions{
    Issuer: "https://resource.example.com",
    AgentServers: []aauth.AgentServer{
        {Issuer: "https://agents.example.com", JwksURI: "https://agents.example.com/jwks.json"},
    },
    AllowedSignatureKeySchemes: []string{"jwt", "jwks_uri", "hwk"},
    SignatureWindow:             60 * time.Second,
}

result := aauth.Verify(ctx, opts, r.Method, r.Host, r.URL.Path, r.Header, jwksClient)

switch result.Identity.Level {
case aauth.LevelAuthorized:
    // pass through
case aauth.LevelIdentified, aauth.LevelPseudonymous:
    // issue a challenge
    challenge := aauth.NewChallenge(challengeOpts, result.Err, result.Identity.AgentHint, false)
    cr := challenge.Build()
    for k, vs := range cr.Headers {
        for _, v := range vs {
            w.Header().Add(k, v)
        }
    }
    w.WriteHeader(cr.StatusCode)
}
```

### Resource server: issuing a resource token

```go
token, err := aauth.MintResourceToken(
    aauth.MintResourceTokenOptions{
        Issuer:     "https://resource.example.com",
        SigningKey:  privateKey,
        KeyID:      "res-key-1",
        TTL:        5 * time.Minute,
    },
    aauth.ResourceTokenClaims{
        Subject:  identity.AgentID,
        Audience: []string{"https://auth.example.com"},
        Scope:    []string{"read"},
    },
)
```

### Agent: signing an outbound request

```go
import "github.com/christian-posta/aauth-go-library/pkg/aauth/agent"

signer, err := agent.NewRequestSigner(agent.SignerOptions{
    AgentID:   "aauth:my-agent@agents.example.com",
    KeyID:     "agent-key-1",
    Signer:    ed25519PrivateKey,
    Algorithm: "ed25519",
    Tokens:    &agent.TokenStore{AgentToken: agentJWT},
})

components := []string{"@method", "@authority", "@path", "signature-key"}
if err := signer.Sign(ctx, req, components); err != nil {
    // handle
}
```

### Agent: performing a full token exchange

```go
authToken, err := agent.ExchangeResourceToken(ctx, agent.ExchangeOptions{
    ResourceToken: resourceToken,
    AuthServerURL: "https://auth.example.com/token",
    HTTPClient:    http.DefaultClient,
})
```

### Agent: transparent signing via RoundTripper

```go
import "github.com/christian-posta/aauth-go-library/pkg/aauth/transport"

client := &http.Client{
    Transport: transport.NewSigningTransport(http.DefaultTransport, signer, components),
}
```

## Testing

```bash
go test ./...
```

For use in tests, `pkg/aauth/aauthtest` provides a `MockJWKSClient` that implements `aauth.JWKSFetcher`:

```go
import "github.com/christian-posta/aauth-go-library/pkg/aauth/aauthtest"

mock := aauthtest.NewMockClient()
mock.Keysets["https://agents.example.com/jwks.json"] = myJWKSet
result := aauth.Verify(ctx, opts, "GET", "resource.example.com", "/api", headers, mock)
```

## Requirements

- Go 1.24+
- [`github.com/lestrrat-go/jwx/v2`](https://github.com/lestrrat-go/jwx) for JWT/JWK handling
