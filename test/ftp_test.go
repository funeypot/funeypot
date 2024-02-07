// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/wolfogre/funeypot/internal/app/config"

	"github.com/jarcoal/httpmock"
	"github.com/jlaffaye/ftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFtpServer(t *testing.T) {
	defer PrepareServers(t, func(cfg *config.Config) {
		cfg.Ftp.Enabled = true
	})()

	t.Run("anonymous", func(t *testing.T) {
		client, err := ftp.Dial("127.0.0.1:2121")
		require.NoError(t, err)

		defer client.Quit()

		err = client.Login("anonymous", "anonymous")
		require.ErrorContains(t, err, "530 Authentication problem: invalid user or password")
	})

	t.Run("auth", func(t *testing.T) {
		client, err := ftp.Dial("127.0.0.1:2121")
		require.NoError(t, err)
		defer client.Quit()

		err = client.Login("username", "password")
		require.ErrorContains(t, err, "530 Authentication problem: invalid user or password")
	})
}

func TestFtpServer_Report(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	defer PrepareServers(t, func(cfg *config.Config) {
		cfg.Ftp.Enabled = true
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
				assert.Equal(t, "18,5", request.Form.Get("categories"))
				assert.Equal(t, `Funeypot detected 5 ftp attempts in 0s. Last by user "username4", password "pas***rd4", client "".`, request.Form.Get("comment"))
				timestamp, _ := time.Parse(time.RFC3339, request.Form.Get("timestamp"))
				assert.WithinDuration(t, time.Now(), timestamp, time.Second)
				return httpmock.NewStringResponse(200, `{}`), nil
			})

		for i := 0; i < 5; i++ {
			client, err := ftp.Dial("127.0.0.1:2121")
			require.NoError(t, err)
			_ = client.Login(fmt.Sprintf("username%d", i), fmt.Sprintf("password%d", i))
			_ = client.Quit()
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
				assert.Equal(t, "18,5", request.Form.Get("categories"))
				assert.Equal(t, `Funeypot detected 6 ftp attempts in 0s. Last by user "username", password "pa****rd", client "".`, request.Form.Get("comment"))
				timestamp, _ := time.Parse(time.RFC3339, request.Form.Get("timestamp"))
				assert.WithinDuration(t, time.Now(), timestamp, time.Second)
				return httpmock.NewStringResponse(200, `{}`), nil
			})

		client, err := ftp.Dial("127.0.0.1:2121")
		require.NoError(t, err)
		defer client.Quit()

		_ = client.Login("username", "password")

		WaitAssert(time.Second, func() bool {
			return httpmock.GetTotalCallCount() > 0
		})
		assert.Equal(t, 1, httpmock.GetTotalCallCount())
	})
}
