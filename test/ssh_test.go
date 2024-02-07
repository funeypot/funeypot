// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package test

import (
	"net"
	"testing"
	"time"

	"github.com/wolfogre/funeypot/internal/app/config"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

func TestSshServer(t *testing.T) {
	defer PrepareServers(t, func(cfg *config.Config) {
		cfg.Ssh.Address = ":2222"
	})()

	sshConfig := &ssh.ClientConfig{
		User:            "username",
		Auth:            []ssh.AuthMethod{ssh.Password("password")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	start := time.Now()
	_, err := ssh.Dial("tcp", "127.0.0.1:2222", sshConfig)
	assert.ErrorContains(t, err, "ssh: handshake failed: ssh: unable to authenticate")
	assert.Greater(t, time.Since(start), 2*time.Second)
}

func TestSshServer_FixedKey(t *testing.T) {
	defer PrepareServers(t, func(cfg *config.Config) {
		cfg.Ssh.Address = ":2222"
		cfg.Ssh.KeySeed = "test"
	})()

	sshConfig := &ssh.ClientConfig{
		User: "username",
		Auth: []ssh.AuthMethod{ssh.Password("password")},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			assert.Equal(t, "SHA256:2HNubLBfVg7yz29A8MBN1wWgRf2tEggxySxBKtm9hJ4", ssh.FingerprintSHA256(key))
			return nil
		},
	}

	start := time.Now()
	_, err := ssh.Dial("tcp", "127.0.0.1:2222", sshConfig)
	assert.ErrorContains(t, err, "ssh: handshake failed: ssh: unable to authenticate")
	assert.Greater(t, time.Since(start), 2*time.Second)
}
