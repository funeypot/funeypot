// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package test

import (
	"net/http"
	"testing"

	"github.com/funeypot/funeypot/internal/app/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboard(t *testing.T) {
	defer PrepareServers(t, func(cfg *config.Config) {
		cfg.Http.Enabled = true
		cfg.Dashboard.Enabled = true
		cfg.Dashboard.Username = "dashboard_username"
		cfg.Dashboard.Password = "dashboard_password"
	})()

	t.Run("no auth", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://127.0.0.1:8080", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("wrong auth", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://127.0.0.1:8080", nil)
		require.NoError(t, err)
		req.SetBasicAuth("username", "password")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("wrong method", func(t *testing.T) {
		t.Run("/api/v1/self", func(t *testing.T) {
			req, err := http.NewRequest("POST", "http://127.0.0.1:8080/api/v1/self", nil)
			require.NoError(t, err)
			req.SetBasicAuth("dashboard_username", "dashboard_password")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			assert.Equal(t, 404, resp.StatusCode)
		})
		t.Run("/api/v1/points", func(t *testing.T) {
			req, err := http.NewRequest("POST", "http://127.0.0.1:8080/api/v1/points", nil)
			require.NoError(t, err)
			req.SetBasicAuth("dashboard_username", "dashboard_password")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			assert.Equal(t, 404, resp.StatusCode)
		})
	})

	t.Run("auth", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://127.0.0.1:8080", nil)
		require.NoError(t, err)
		req.SetBasicAuth("dashboard_username", "dashboard_password")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("/api/v1/self", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://127.0.0.1:8080/api/v1/self", nil)
		require.NoError(t, err)
		req.SetBasicAuth("dashboard_username", "dashboard_password")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("/api/v1/points", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://127.0.0.1:8080/api/v1/points", nil)
		require.NoError(t, err)
		req.SetBasicAuth("dashboard_username", "dashboard_password")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}
