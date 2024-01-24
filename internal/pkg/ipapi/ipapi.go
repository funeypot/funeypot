package ipapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

/*
TODO: check
      < X-Ttl: 3
      < X-Rl: 34
*/

func Query(ctx context.Context, ip string) (*Response, error) {
	result := &Response{}

	_, err := resty.New().R().
		SetContext(ctx).
		SetResult(result).
		Get(fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,continent,continentCode,country,countryCode,region,regionName,city,district,zip,lat,lon,timezone,offset,currency,isp,org,as,asname,reverse,mobile,proxy,hosting,query", ip))
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("%s: %s", result.Status, result.Message)
	}

	return result, nil
}

type Response struct {
	Query         string  `json:"query"`
	Status        string  `json:"status"`
	Message       string  `json:"message"`
	Continent     string  `json:"continent"`
	ContinentCode string  `json:"continentCode"`
	Country       string  `json:"country"`
	CountryCode   string  `json:"countryCode"`
	Region        string  `json:"region"`
	RegionName    string  `json:"regionName"`
	City          string  `json:"city"`
	District      string  `json:"district"`
	Zip           string  `json:"zip"`
	Lat           float64 `json:"lat"`
	Lon           float64 `json:"lon"`
	Timezone      string  `json:"timezone"`
	Offset        int     `json:"offset"`
	Currency      string  `json:"currency"`
	Isp           string  `json:"isp"`
	Org           string  `json:"org"`
	As            string  `json:"as"`
	Asname        string  `json:"asname"`
	Reverse       string  `json:"reverse"`
	Mobile        bool    `json:"mobile"`
	Proxy         bool    `json:"proxy"`
	Hosting       bool    `json:"hosting"`
}

func (r *Response) Location() string {
	strs := make([]string, 0, 4)
	if r.Country != "" {
		strs = append(strs, r.Country)
	}
	if r.RegionName != "" {
		strs = append(strs, r.RegionName)
	}
	if r.City != "" {
		strs = append(strs, r.City)
	}
	if r.District != "" {
		strs = append(strs, r.District)
	}
	return strings.Join(strs, ", ")
}
