// This file was generated by typeinst. Do not edit, use `go generate` instead.
package main

import (
	ast "go/ast"
)

type StrSet map[string]struct{}

func NewStrSet() StrSet {
	return make(map[string]struct{})
}

func (set StrSet) Add(v string) {
	set[v] = struct{}{}
}

func (set StrSet) Has(v string) bool {
	_, has := set[v]
	return has
}

type AstIdentSet map[*ast.Ident]struct{}

func NewAstIdentSet() AstIdentSet {
	return make(map[*ast.Ident]struct{})
}

func (set AstIdentSet) Add(v *ast.Ident) {
	set[v] = struct{}{}
}

func (set AstIdentSet) Has(v *ast.Ident) bool {
	_, has := set[v]
	return has
}
