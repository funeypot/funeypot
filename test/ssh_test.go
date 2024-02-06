// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

func TestSshServer(t *testing.T) {
	defer PrepareServers(t, nil)()

	config := &ssh.ClientConfig{
		User:            "username",
		Auth:            []ssh.AuthMethod{ssh.Password("password")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	start := time.Now()
	_, err := ssh.Dial("tcp", "127.0.0.1:22", config)
	assert.ErrorContains(t, err, "ssh: handshake failed: ssh: unable to authenticate")
	assert.Greater(t, time.Since(start), 2*time.Second)
}
