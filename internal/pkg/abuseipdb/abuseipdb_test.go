// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package abuseipdb

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestClient_Report(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	client := NewClient("test_key", time.Second)

	t.Run("Cooldown", func(t *testing.T) {
		defer httpmock.Reset()
		httpmock.RegisterResponder("POST", ReportUrl,
			func(request *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(429, `{
  "errors": [
      {
          "detail": "Daily rate limit of 1000 requests exceeded for this endpoint. See headers for additional details.",
          "status": 429
      }
  ]
}`)
				resp.Header.Set("Retry-After", "60")
				return resp, nil
			})
		until, ok := client.Cooldown()
		assert.False(t, ok)
		assert.Zero(t, until)

		score, err := client.ReportSsh(context.Background(), "127.0.0.1", time.Now(), "test")
		assert.ErrorContains(t, err, "429")
		assert.Zero(t, score)

		until, ok = client.Cooldown()
		assert.True(t, ok)
		assert.WithinDuration(t, time.Now().Add(60*time.Second), until, 2*time.Second)
	})
}
