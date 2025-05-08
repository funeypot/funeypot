// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package sshkey

import (
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateKey(t *testing.T) {
	t.Run("empty seed", func(t *testing.T) {
		hashM := make(map[string]bool)
		for i := 0; i < 10; i++ {
			key, err := GenerateKey("")
			require.NoError(t, err)
			hash := privateKeyHash(key)
			require.False(t, hashM[hash])
			hashM[hash] = true
		}
	})

	t.Run("same seed", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			key, err := GenerateKey("1")
			require.NoError(t, err)
			require.Equal(t, "ff3457080d98fed59a8356958690c014", privateKeyHash(key))
		}
	})

	t.Run("another seed", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			key, err := GenerateKey("1234")
			require.NoError(t, err)
			require.Equal(t, "9d6d0b4a76570b69c77188a1e4715df4", privateKeyHash(key))
		}
	})
}

func privateKeyHash(key *rsa.PrivateKey) string {
	sum := md5.Sum(x509.MarshalPKCS1PrivateKey(key))
	return hex.EncodeToString(sum[:])
}
