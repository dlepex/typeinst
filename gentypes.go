package main

import (
	"go/ast"

	"github.com/dlepex/typeinst/generic/set"
)

//go:generate typeinst
type _typeinst struct {
	StrSet      func(E string) set.Set
	AstIdentSet func(E *ast.Ident) set.Set
}
