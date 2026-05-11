# aauth-go-library

Go library for the AAuth protocol — authentication and authorization for autonomous agents.

Used in production by [extauth-aauth-resource](https://github.com/christian-posta/extauth-aauth-resource), an [Agentgateway](https://agentgateway.dev)/Envoy external authorization filter ([ext_authz](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authz/v3/ext_authz.proto)) that enforces [AAuth](https://www.aauth.dev) on inbound service requests.

## What it does

[AAuth](https://www.aauth.dev) is an auth protocol that defines how an AI agent proves its identity and obtains authorization to call a protected resource. This library handles both sides of that exchange (caller/receiver).

**Resource servers** get primitives to verify inbound agent requests, issue 401 challenges when authorization is missing or insufficient, and mint short-lived resource tokens that agents can exchange for auth tokens.

**Agents** get an SDK for signing outbound HTTP requests (RFC 9421), handling 401 challenges, exchanging resource tokens with an authorization server, and polling for deferred 202 responses.

## Installation

```bash
go get github.com/christian-posta/aauth-go-library
```

## Usage

### Resource server: verifying an inbound request

```go
import "github.com/christian-posta/aauth-go-library/pkg/aauth"

opts := aauth.VerifyOptions{
    Issuer: "https://resource.example.com",
    AgentServers: []aauth.AgentServer{
        {Issuer: "https://agents.example.com", JwksURI: "https://agents.example.com/jwks.json"},
    },
    AllowedSignatureKeySchemes: []string{"jwt", "jwks_uri", "hwk"},
    AllowedJWTTypes:            []string{"aa-agent+jwt", "aa-auth+jwt"},
    SignatureWindow:             60 * time.Second,
}

result := aauth.Verify(ctx, opts, r.Method, r.Host, r.URL.Path, r.Header, jwksClient)

switch result.Identity.Level {
case aauth.LevelAuthorized:
    // pass through
case aauth.LevelIdentified, aauth.LevelPseudonymous:
    challenge := aauth.NewChallenge(challengeOpts, result.Err, nil, false)
    cr := challenge.Build()
    for k, vs := range cr.Headers {
        for _, v := range vs {
            w.Header().Add(k, v)
        }
    }
    w.WriteHeader(cr.Status)
}
```

### Resource server: minting a resource token

```go
token, err := aauth.MintResourceToken(
    aauth.MintResourceTokenOptions{
        Issuer:        "https://resource.example.com",
        Aud:           "https://auth.example.com",
        SigningKeyKid: "res-key-1",
        SigningKey:    privateKey,
    },
    aauth.ResourceTokenClaims{
        Iss:      "https://resource.example.com",
        Agent:    "aauth:alice@agents.example.com",
        AgentJKT: agentThumbprint,
        Exp:      time.Now().Add(5 * time.Minute).Unix(),
        Scope:    "read",
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

`ExchangeResourceToken` wires together the challenge handler, token exchanger, and deferred poller into a single call.

```go
result, err := agent.ExchangeResourceToken(ctx, agent.ExchangeResourceTokenOptions{
    ResourceToken: resourceToken,
    Signer:        signer,
    PSMetadataURL: "https://person.example.com/.well-known/aauth-person.json",
    HTTPClient:    http.DefaultClient,
})
// result.AuthToken is the bearer token to use on the resource request
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

`pkg/aauth/aauthtest` provides a `MockJWKSClient` that implements `aauth.JWKSFetcher`:

```go
import "github.com/christian-posta/aauth-go-library/pkg/aauth/aauthtest"

mock := aauthtest.NewMockClient()
mock.Keysets["https://agents.example.com/jwks.json"] = myJWKSet
result := aauth.Verify(ctx, opts, "GET", "resource.example.com", "/api", headers, mock)
```

## Requirements

- Go 1.24+
- [`github.com/lestrrat-go/jwx/v2`](https://github.com/lestrrat-go/jwx) for JWT/JWK handling
