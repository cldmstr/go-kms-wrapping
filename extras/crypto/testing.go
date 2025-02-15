package crypto

import (
	"context"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"testing"

	wrapping "github.com/hashicorp/go-kms-wrapping/v2"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"
)

// TestWithEd25519 produces test hmac sha256 using a derived Ed25519 key
func TestWithEd25519(t *testing.T, data []byte, w wrapping.Wrapper, opt ...wrapping.Option) string {
	t.Helper()
	require := require.New(t)
	reader, err := NewDerivedReader(context.Background(), w, 32, opt...)
	require.NoError(err)
	edKey, _, err := ed25519.GenerateKey(reader)
	require.NoError(err)
	var key [32]byte
	n := copy(key[:], edKey)
	require.Equal(n, 32)
	return TestHmacSha256(t, key[:], data, opt...)
}

// TestWithBlake2b produces a test hmac sha256 using derived blake2b.  Supported
// options: WithPrk
func TestWithBlake2b(t *testing.T, data []byte, w wrapping.Wrapper, opt ...wrapping.Option) string {
	t.Helper()
	require := require.New(t)
	require.NotNil(data)
	require.NotNil(w)
	opts, err := getOpts(opt...)
	require.NoError(err)
	var key [32]byte
	switch {
	case opts.withPrk != nil:
		key = blake2b.Sum256(opts.withPrk)
	default:
		reader, err := NewDerivedReader(context.Background(), w, 32, opt...)
		require.NoError(err)
		readerKey := make([]byte, 32)
		n, err := io.ReadFull(reader, readerKey)
		require.NoError(err)
		require.Equal(n, 32)
		key = blake2b.Sum256(readerKey)
	}
	return TestHmacSha256(t, key[:], data, opt...)
}

// TestHmacSha256 produces a test hmac sha256
func TestHmacSha256(t *testing.T, key, data []byte, opt ...wrapping.Option) string {
	t.Helper()
	require := require.New(t)
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	hmac := mac.Sum(nil)
	var hmacString string
	opts, err := getOpts(opt...)
	require.NoError(err)
	switch {
	case opts.withBase64Encoding:
		hmacString = base64.RawURLEncoding.EncodeToString(hmac)
	case opts.withBase58Encoding:
		hmacString = base58.Encode(hmac)
	default:
		hmacString = string(hmac)
	}
	if opts.withPrefix != "" {
		return opts.withPrefix + hmacString
	}
	return hmacString
}
