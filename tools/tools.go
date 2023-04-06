//go:build tools
// +build tools

package tools

import (
	_ "github.com/99designs/gqlgen"
	_ "github.com/amacneil/dbmate"
	_ "github.com/kyleconroy/sqlc/cmd/sqlc"
)