package fakever

import (
	_ "embed"
)

//go:embed ssh.txt
var SshVersion string
