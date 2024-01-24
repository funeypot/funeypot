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
			require.Equal(t, "2a7b14dfdb31f6ae43c7427d74c9b133", privateKeyHash(key))
		}
	})

	t.Run("another seed", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			key, err := GenerateKey("1234")
			require.NoError(t, err)
			require.Equal(t, "fdfa43f717bbeb688e80c48cfacdb2f4", privateKeyHash(key))
		}
	})
}

func privateKeyHash(key *rsa.PrivateKey) string {
	sum := md5.Sum(x509.MarshalPKCS1PrivateKey(key))
	return hex.EncodeToString(sum[:])
}
