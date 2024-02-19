// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/funeypot/funeypot/internal/app/config"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpServer(t *testing.T) {
	defer PrepareServers(t, func(cfg *config.Config) {
		cfg.Http.Enabled = true
	})()

	t.Run("no auth", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:8080")
		require.NoError(t, err)

		assert.Equal(t, 401, resp.StatusCode)
		assert.Equal(t, `Basic realm="Restricted"`, resp.Header.Get("WWW-Authenticate"))
	})

	t.Run("auth", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://127.0.0.1:8080", nil)
		require.NoError(t, err)
		req.SetBasicAuth("username", "password")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode)
		assert.Equal(t, `Basic realm="Restricted"`, resp.Header.Get("WWW-Authenticate"))
	})
}

func TestHttpServer_Report(t *testing.T) {
	HttpClient := &http.Client{
		Transport: http.DefaultTransport,
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	defer PrepareServers(t, func(cfg *config.Config) {
		cfg.Http.Enabled = true
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
				assert.Equal(t, "18,21", request.Form.Get("categories"))
				assert.Equal(t, `Funeypot detected 5 http attempts in 0s. Last by user "username4", password "pas***rd4", client "Go-http-client/1.1".`, request.Form.Get("comment"))
				timestamp, _ := time.Parse(time.RFC3339, request.Form.Get("timestamp"))
				assert.WithinDuration(t, time.Now(), timestamp, 2*time.Second)
				return httpmock.NewStringResponse(200, `{}`), nil
			})

		for i := 0; i < 5; i++ {
			req, err := http.NewRequest("GET", "http://127.0.0.1:8080", nil)
			require.NoError(t, err)
			req.SetBasicAuth(fmt.Sprintf("username%d", i), fmt.Sprintf("password%d", i))
			_, _ = HttpClient.Do(req)
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
				assert.Equal(t, "18,21", request.Form.Get("categories"))
				assert.Equal(t, `Funeypot detected 6 http attempts in 0s. Last by user "username", password "pa****rd", client "Go-http-client/1.1".`, request.Form.Get("comment"))
				timestamp, _ := time.Parse(time.RFC3339, request.Form.Get("timestamp"))
				assert.WithinDuration(t, time.Now(), timestamp, 2*time.Second)
				return httpmock.NewStringResponse(200, `{}`), nil
			})

		req, err := http.NewRequest("GET", "http://127.0.0.1:8080", nil)
		require.NoError(t, err)
		req.SetBasicAuth("username", "password")
		_, _ = HttpClient.Do(req)

		WaitAssert(time.Second, func() bool {
			return httpmock.GetTotalCallCount() > 0
		})
		assert.Equal(t, 1, httpmock.GetTotalCallCount())
	})
}
