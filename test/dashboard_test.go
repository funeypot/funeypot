// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package test

import (
	"net/http"
	"testing"

	"github.com/wolfogre/funeypot/internal/app/config"

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

	t.Run("view", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://127.0.0.1:8080", nil)
		require.NoError(t, err)
		req.SetBasicAuth("dashboard_username", "dashboard_password")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}
