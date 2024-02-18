// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package ipgeo

import (
	"context"
	"net"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func TestIpapiComQuerier_Query(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://ip-api.com/json/2.3.4.5?fields=status,message,continent,continentCode,country,countryCode,region,regionName,city,district,zip,lat,lon,timezone,offset,currency,isp,org,as,asname,reverse,mobile,proxy,hosting,query",
		func(request *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, `
{
  "status": "success",
  "continent": "Europe",
  "continentCode": "EU",
  "country": "France",
  "countryCode": "FR",
  "region": "ARA",
  "regionName": "Auvergne-Rhone-Alpes",
  "city": "Clermont-Ferrand",
  "district": "",
  "zip": "63000",
  "lat": 45.7838,
  "lon": 3.0966,
  "timezone": "Europe/Paris",
  "offset": 3600,
  "currency": "EUR",
  "isp": "France Telecom Orange",
  "org": "",
  "as": "AS3215 Orange S.A.",
  "asname": "AS3215",
  "reverse": "lfbn-cle-1-191-5.w2-3.abo.wanadoo.fr",
  "mobile": false,
  "proxy": false,
  "hosting": false,
  "query": "2.3.4.5"
}
`)
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		})

	t.Run("regular", func(t *testing.T) {
		querier := NewIpapiComQuerier()
		info, err := querier.Query(context.Background(), "2.3.4.5")
		require.NoError(t, err)
		require.Equal(t, &Info{
			Ip:        net.ParseIP("2.3.4.5"),
			Location:  "France, Auvergne-Rhone-Alpes, Clermont-Ferrand",
			Latitude:  45.7838,
			Longitude: 3.0966,
		}, info)
	})
}
