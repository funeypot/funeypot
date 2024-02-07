// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package test

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/wolfogre/funeypot/internal/app/config"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

func TestSshServer(t *testing.T) {
	defer PrepareServers(t, nil)()

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

func TestSshServer_Report(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	defer PrepareServers(t, func(cfg *config.Config) {
		cfg.Abuseipdb.Enabled = true
		cfg.Abuseipdb.Key = "test_key"
		cfg.Abuseipdb.Interval = 0
	})()

	t.Run("first", func(t *testing.T) {
		defer httpmock.Reset()
		httpmock.RegisterResponder("POST", "https://api.abuseipdb.com/api/v2/report",
			func(request *http.Request) (*http.Response, error) {
				assert.Equal(t, "test_key", request.Header.Get("Key"))
				assert.Equal(t, "application/x-www-form-urlencoded", request.Header.Get("Content-Type"))
				assert.NoError(t, request.ParseForm())
				assert.Equal(t, "127.0.0.1", request.Form.Get("ip"))
				assert.Equal(t, "18,22", request.Form.Get("categories"))
				assert.Equal(t, `Funeypot detected 5 ssh attempts in 8s. Last by user "username", password "pa****rd", client "Go".`, request.Form.Get("comment"))
				timestamp, _ := time.Parse(time.RFC3339, request.Form.Get("timestamp"))
				assert.WithinDuration(t, time.Now(), timestamp, 2*time.Second)
				return httpmock.NewStringResponse(200, `{}`), nil
			})

		sshConfig := &ssh.ClientConfig{
			User:            "username",
			Auth:            []ssh.AuthMethod{ssh.Password("password")},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		for i := 0; i < 5; i++ {
			_, _ = ssh.Dial("tcp", "127.0.0.1:2222", sshConfig)
		}
		WaitAssert(time.Second, func() bool {
			return httpmock.GetTotalCallCount() > 0
		})
		assert.Equal(t, 1, httpmock.GetTotalCallCount())
	})

	t.Run("continue", func(t *testing.T) {
		defer httpmock.Reset()
		httpmock.RegisterResponder("POST", "https://api.abuseipdb.com/api/v2/report",
			func(request *http.Request) (*http.Response, error) {
				assert.Equal(t, "test_key", request.Header.Get("Key"))
				assert.Equal(t, "application/x-www-form-urlencoded", request.Header.Get("Content-Type"))
				assert.NoError(t, request.ParseForm())
				assert.Equal(t, "127.0.0.1", request.Form.Get("ip"))
				assert.Equal(t, "18,22", request.Form.Get("categories"))
				assert.Equal(t, `Funeypot detected 6 ssh attempts in 10s. Last by user "username2", password "pas***rd2", client "Go".`, request.Form.Get("comment"))
				timestamp, _ := time.Parse(time.RFC3339, request.Form.Get("timestamp"))
				assert.WithinDuration(t, time.Now(), timestamp, 2*time.Second)
				return httpmock.NewStringResponse(200, `{}`), nil
			})

		sshConfig := &ssh.ClientConfig{
			User:            "username2",
			Auth:            []ssh.AuthMethod{ssh.Password("password2")},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		_, _ = ssh.Dial("tcp", "127.0.0.1:2222", sshConfig)

		WaitAssert(time.Second, func() bool {
			return httpmock.GetTotalCallCount() > 0
		})
		assert.Equal(t, 1, httpmock.GetTotalCallCount())
	})
}
