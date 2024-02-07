// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package ipapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponse_Location(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		want     string
	}{
		{
			name: "regular",
			response: Response{
				Country:    "A",
				RegionName: "B",
				City:       "C",
				District:   "D",
			},
			want: "A, B, C, D",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.response.Location())
		})
	}
}

func TestQuery(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	t.Run("regular", func(t *testing.T) {
		defer httpmock.Reset()
		httpmock.RegisterResponder("GET", "http://ip-api.com/json/123.45.67.89?fields=status,message,continent,continentCode,country,countryCode,region,regionName,city,district,zip,lat,lon,timezone,offset,currency,isp,org,as,asname,reverse,mobile,proxy,hosting,query",
			func(request *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, `
{
  "status": "success",
  "message": "Test",
  "continent": "Asia",
  "continentCode": "AS",
  "country": "South Korea",
  "countryCode": "KR",
  "region": "11",
  "regionName": "Seoul",
  "city": "Seongnam-si",
  "district": "",
  "zip": "05670",
  "lat": 37.5161,
  "lon": 127.1,
  "timezone": "Asia/Seoul",
  "offset": 32400,
  "currency": "KRW",
  "isp": "SamsungSDS Inc",
  "org": "SamsungSDS Inc",
  "as": "",
  "asname": "",
  "reverse": "",
  "mobile": false,
  "proxy": true,
  "hosting": false,
  "query": "123.45.67.89"
}
`)
				resp.Header.Set("Content-Type", "application/json")
				return resp, nil
			})
		resp, err := Query(context.Background(), "123.45.67.89")
		require.NoError(t, err)
		assert.Equal(t, &Response{
			Query:         "123.45.67.89",
			Status:        "success",
			Message:       "Test",
			Continent:     "Asia",
			ContinentCode: "AS",
			Country:       "South Korea",
			CountryCode:   "KR",
			Region:        "11",
			RegionName:    "Seoul",
			City:          "Seongnam-si",
			District:      "",
			Zip:           "05670",
			Lat:           37.5161,
			Lon:           127.1,
			Timezone:      "Asia/Seoul",
			Offset:        32400,
			Currency:      "KRW",
			Isp:           "SamsungSDS Inc",
			Org:           "SamsungSDS Inc",
			As:            "",
			Asname:        "",
			Reverse:       "",
			Mobile:        false,
			Proxy:         true,
			Hosting:       false,
		}, resp)
	})
}
