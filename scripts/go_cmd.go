//go:build exclude

package scripts

// To make sure the wire command is available in Makefile.
import _ "github.com/google/wire/cmd/wire"

// To make sure the stringer command is available in Makefile.
import _ "golang.org/x/tools/cmd/stringer"
