package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strings"
)

// Imports is name-path bimap
type Imports struct {
	p2n map[string]string // path -> name, one side of bimap
	n2p map[string]string // the other side
}

func (im *Imports) init() {
	if im.p2n == nil {
		im.p2n = make(map[string]string)
		im.n2p = make(map[string]string)
	}
}

// IsEmpty -
func (im *Imports) IsEmpty() bool { return len(im.p2n) == 0 }

// AddSpec -
func (im *Imports) AddSpec(spec *ast.ImportSpec) error {
	return im.Add(importSpecName(spec), spec.Path.Value)
}

func importSpecName(spec *ast.ImportSpec) string {
	if spec.Name != nil {
		return spec.Name.Name
	}
	a := strings.FieldsFunc(spec.Path.Value, func(r rune) bool {
		return r == '"' || r == '/'
	})
	if len(a) == 0 {
		log.Fatal("Nameless import spec")
	}
	return a[len(a)-1]

}

// Add n - import name, p - import path
func (im *Imports) Add(n, p string) error {
	if n == "." {
		return fmt.Errorf("dot-import is not allowed: %s", p)
	}
	if n == "" {
		return fmt.Errorf("empty import is not allowed: %s", p)
	}
	im.init()
	if oldn, ok := im.p2n[p]; ok {
		if oldn != n {
			return fmt.Errorf("import package %s under different names: %s, %s  ", p, n, oldn)
		}
		return nil
	}
	if oldp, ok := im.n2p[n]; ok {
		if oldp != p {
			return fmt.Errorf("import package: %s under name: '%s' that was already used for other package: %s", p, n, oldp)
		}
	}
	im.n2p[n] = p
	im.p2n[p] = n
	return nil
}

// Merge keeps old import names, renaming happens in new ("other") imports
// Merge returns a rename map.
func (im *Imports) Merge(other Imports) map[string]string {
	add := [][2]string{}
	rename := make(map[string]string)
	for n, p := range other.n2p {
		oldn, hasp := im.p2n[p]
		_, hasn := im.n2p[n]
		if hasp {
			if oldn != n {
				rename[n] = oldn
			}
		} else {
			if !hasn {
				add = append(add, [2]string{n, p})
			} else {
				gs := genSymbol("_Pkg")
				rename[n] = gs
				add = append(add, [2]string{gs, p})
			}
		}
	}
	for _, pair := range add {
		_ = im.Add(pair[0], pair[1])
	}
	return rename
}

func (im *Imports) decl() *ast.GenDecl {
	specs := make([]ast.Spec, 0, len(im.n2p))
	for n, p := range im.n2p {
		spec := &ast.ImportSpec{
			Name: &ast.Ident{
				Name: n,
				Obj: &ast.Object{
					Kind: ast.Pkg,
				},
			},
			Path: &ast.BasicLit{
				Value: p,
				Kind:  token.STRING,
			},
		}
		specs = append(specs, spec)
	}

	return &ast.GenDecl{
		Lparen: 1,
		Rparen: 1,
		Specs:  specs,
		Tok:    token.IMPORT,
	}
}

//Named returns import path by name
func (im *Imports) Named(n string) string {
	return im.n2p[n]
}

func (im *Imports) requireNamed(n string) string {
	p := im.Named(n)
	if p == "" {
		bpan.Panicf("unresolved import named: %s", n)
	}
	return p
}
