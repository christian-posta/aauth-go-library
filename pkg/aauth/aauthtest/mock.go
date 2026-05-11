package aauthtest

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

// MockJWKSClient is a test double for aauth.JWKSFetcher.
type MockJWKSClient struct {
	Keysets         map[string]jwk.Set
	Metadata        map[string]map[string]interface{}
	GetCalls        map[string]int
	InvalidateCalls map[string]int
	OnInvalidate    func(uri string)
}

func NewMockClient() *MockJWKSClient {
	return &MockJWKSClient{
		Keysets:         make(map[string]jwk.Set),
		Metadata:        make(map[string]map[string]interface{}),
		GetCalls:        make(map[string]int),
		InvalidateCalls: make(map[string]int),
	}
}

func (m *MockJWKSClient) Get(ctx context.Context, uri string) (jwk.Set, error) {
	m.GetCalls[uri]++
	if set, ok := m.Keysets[uri]; ok {
		return set, nil
	}
	return nil, fmt.Errorf("mock: not found")
}

func (m *MockJWKSClient) GetMetadata(ctx context.Context, uri string) (map[string]interface{}, error) {
	if md, ok := m.Metadata[uri]; ok {
		return md, nil
	}
	return nil, fmt.Errorf("mock metadata: not found")
}

func (m *MockJWKSClient) Invalidate(uri string) {
	m.InvalidateCalls[uri]++
	if m.OnInvalidate != nil {
		m.OnInvalidate(uri)
	}
}
