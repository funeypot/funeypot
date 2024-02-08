// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package ipgeo

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMaxmindQuerier(t *testing.T) {
	t.Run("embed", func(t *testing.T) {
		querier, err := NewMaxmindQuerier("embed")
		require.NoError(t, err)

		t.Run("query", func(t *testing.T) {
			info, err := querier.Query(context.Background(), "2.3.4.5")
			require.NoError(t, err)
			assert.Equal(t, &Info{
				Ip:        net.ParseIP("2.3.4.5"),
				Location:  "France, Auvergne-Rhone-Alpes, Puy-de-Dôme, Clermont-Ferrand",
				Latitude:  45.7838,
				Longitude: 3.0966,
			}, info)
		})

		t.Run("query invalid ip", func(t *testing.T) {
			_, err := querier.Query(context.Background(), "2.3.4.5.6")
			assert.ErrorContains(t, err, "invalid ip address")
		})
	})

	t.Run("release again", func(t *testing.T) {
		_, _ = NewMaxmindQuerier("embed")
		_, err := NewMaxmindQuerier("embed")
		require.NoError(t, err)
	})

	t.Run("use no embed", func(t *testing.T) {
		_, err := NewMaxmindQuerier("test.mmdb")
		require.ErrorContains(t, err, "no such file or directory")
	})

	t.Run("no embed", func(t *testing.T) {
		geoLite2City = nil
		_, err := NewMaxmindQuerier("embed")
		require.ErrorContains(t, err, "you are running a version without the embedded")
	})
}
