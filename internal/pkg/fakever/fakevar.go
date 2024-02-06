// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package fakever

import (
	_ "embed"
)

//go:embed ssh.txt
var SshVersion string
