// Package aauth implements the resource-side primitives of the AAuth protocol
// (Authenticated Authorization for autonomous agents).
//
// The package is split into a small resource-side surface — Verify, Identity,
// Challenge, MintResourceToken — and a set of focused sub-packages used by both
// resource servers and agent clients:
//
//   - [github.com/christian-posta/aauth-go-library/pkg/aauth/agent]       — agent SDK (signing, exchange, polling)
//   - [github.com/christian-posta/aauth-go-library/pkg/aauth/headers]     — AAuth-Requirement, Accept-Signature, etc.
//   - [github.com/christian-posta/aauth-go-library/pkg/aauth/http]        — 202/Deferred response helpers
//   - [github.com/christian-posta/aauth-go-library/pkg/aauth/identifiers] — AAuth server/agent identifier validation
//   - [github.com/christian-posta/aauth-go-library/pkg/aauth/keys]        — JWKS fetcher with caching, JWK conversion
//   - [github.com/christian-posta/aauth-go-library/pkg/aauth/metadata]    — well-known metadata documents and fetcher
//   - [github.com/christian-posta/aauth-go-library/pkg/aauth/transport]   — http.RoundTripper that signs requests
//
// Resource servers usually call [Verify] inside an extauthz adapter, then build
// a [Challenge] in the unauthenticated case. Implements AAuth SPEC §4 (token
// verification), §6 (signature verification) and §12 (challenge headers).
package aauth
