// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package sshkey

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	mathrand "math/rand"
	"runtime"

	"golang.org/x/crypto/ssh"
)

func GenerateSigner(seed string) (ssh.Signer, error) {
	key, err := rsa.GenerateKey(newRandReader(seed), 2048)
	if err != nil {
		return nil, err
	}
	return ssh.NewSignerFromKey(key)
}

func GenerateKey(seed string) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(newRandReader(seed), 2048)
}

type randReader struct {
	rand *mathrand.Rand
}

func newRandReader(seed string) randReader {
	if seed == "" {
		return randReader{}
	}
	md5sum := sha256.Sum256([]byte(seed))
	return randReader{
		rand: mathrand.New(mathrand.NewSource(int64(binary.LittleEndian.Uint64(md5sum[:8])))),
	}
}

func (r randReader) Read(p []byte) (n int, err error) {
	if r.rand == nil {
		return rand.Read(p)
	}

	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	if runtime.FuncForPC(pc[0]-1).Name() == "crypto/internal/randutil.MaybeReadByte" {
		return 1, nil
	}

	for i := range p {
		p[i] = byte(r.rand.Intn(256))
	}
	return len(p), nil
}
