// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

//go:build NO_EMBED_IPGEO

package ipgeo

import (
	_ "embed"
)

var (
	geoLite2City []byte
)
