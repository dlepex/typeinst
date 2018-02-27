package main

import (
	"go/ast"

	"github.com/dlepex/typeinst/generic/set"
)

//go:generate typeinst
type _typeinst struct {
	StrSet      func(T string) set.Set
	AstIdentSet func(T *ast.Ident) set.Set
}
