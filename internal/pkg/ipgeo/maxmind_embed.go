// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

//go:build !NO_EMBED_IPGEO

package ipgeo

import (
	_ "embed"
)

var (
	//go:embed GeoLite2-City.mmdb
	geoLite2City []byte
)
