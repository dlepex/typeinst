package main

import (
	"go/ast"

	"github.com/dlepex/genericlib/set"
)

//go:generate typeinst
type _typeinst struct { // nolint
	StrSet      func(E string) set.Set
	AstIdentSet func(E *ast.Ident) set.Set
}
