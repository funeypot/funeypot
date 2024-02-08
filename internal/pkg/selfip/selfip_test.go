// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package selfip

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUrls(t *testing.T) {
	expected := ""

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			startedAt := time.Now()
			ip, err := getFrom(context.Background(), url)
			duration := time.Since(startedAt)

			require.NoError(t, err)
			if expected == "" {
				expected = ip
			}

			assert.Equal(t, expected, ip)

			t.Logf("%s (%s)", ip, duration)
		})
	}
}

func TestGet(t *testing.T) {
	t.Run("regular", func(t *testing.T) {
		ip, err := Get(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, net.ParseIP(ip))
	})
	t.Run("timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		_, err := Get(ctx)
		assert.ErrorContains(t, err, "context deadline exceeded")
	})
	t.Run("all failed", func(t *testing.T) {
		transport := http.DefaultTransport
		http.DefaultTransport = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 1 * time.Millisecond,
			}).DialContext,
		}
		defer func() {
			http.DefaultTransport = transport
		}()

		_, err := Get(context.Background())
		assert.ErrorContains(t, err, "all attempts failed")
	})
}

func Test_getFrom(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	t.Run("regular", func(t *testing.T) {
		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, "http://test.com",
			httpmock.NewStringResponder(http.StatusOK, "1.2.3.4"))
		ip, err := getFrom(context.Background(), "http://test.com")
		require.NoError(t, err)
		assert.Equal(t, "1.2.3.4", ip)
	})

	t.Run("invalid url", func(t *testing.T) {
		_, err := getFrom(context.Background(), ":")
		assert.ErrorContains(t, err, "new request:")
	})

	t.Run("invalid code", func(t *testing.T) {
		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, "http://test.com",
			httpmock.NewStringResponder(http.StatusBadRequest, "1.2.3.4"))
		_, err := getFrom(context.Background(), "http://test.com")
		require.ErrorContains(t, err, "status: 400")
	})

	t.Run("broken body", func(t *testing.T) {
		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, "http://test.com",
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(http.StatusOK, "")
				resp.Body = errorReadCloser{}
				return resp, nil
			})
		_, err := getFrom(context.Background(), "http://test.com")
		require.ErrorContains(t, err, "mocked IO error")
	})

	t.Run("invalid response", func(t *testing.T) {
		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, "http://test.com",
			httpmock.NewStringResponder(http.StatusOK, ""))
		_, err := getFrom(context.Background(), "http://test.com")
		require.ErrorContains(t, err, `invalid response: ""`)
	})
}

type errorReadCloser struct{}

func (rc errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("mocked IO error")
}

func (rc errorReadCloser) Close() error {
	return nil
}
