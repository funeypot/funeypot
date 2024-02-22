// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"os"

	"github.com/funeypot/funeypot/internal/app/entry"
)

var (
	Version = "dev"
)

func main() {
	// TODO: support more subcommands

	if err := entry.Run(context.Background(), Version, os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
