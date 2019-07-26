package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/token"
	"os"

	pri "github.com/dlepex/typeinst/printer"
)

const preambleComment = `// This file was generated by typeinst. Do not edit, use "go generate" instead.
// nolint`

type astPrinter struct {
	pri.Config
	w    *bufio.Writer
	fset *token.FileSet
}

func newAstPrinter(w *bufio.Writer, rf pri.RenameFunc) *astPrinter {
	return &astPrinter{pri.Config{
		Mode:       pri.UseSpaces,
		RenameFunc: rf,
	}, w, token.NewFileSet()}
}

func (p *astPrinter) println(node interface{}) {
	err := p.Fprint(p.w, p.fset, node)
	if err != nil {
		bpan.Panicf("Print AST error (%v) for node: %v", err, node)
	}
	_, err = p.w.WriteString("\n\n")
	if err != nil {
		bpan.Panicf("Writer error: %v", err)
	}
}

// Print prints impl to file
func (im *Impl) Print() (err error) {
	f, err := os.Create(im.outputFile)
	if err != nil {
		return err
	}
	defer bpan.RecoverTo(&err)
	defer func() { err = f.Close() }()
	wr := bufio.NewWriter(f)
	defer func() { err = wr.Flush() }()
	fmt.Fprintf(wr, "%s\npackage %s\n\n", preambleComment, im.pkgName)
	if !im.imports.IsEmpty() {
		newAstPrinter(wr, nil).println(im.imports.decl())
	}
	typedefs := NewStrSet()
	for _, p := range im.pkg {
		p.print(wr, typedefs)
	}
	return
}

func (td *TypeDesc) printedName(n string) string {
	if !td.isSingleton {
		return n
	}
	// for singleton types their name is used for var-declaration
	return n + "Type"
}

func (pk *PkgDesc) renameFunc(args *TypeArgs, inCtor bool) pri.RenameFunc {

	stringer := astStringer{}
	return func(id *ast.Ident) string {
		n := id.Name
		if pk.occTypes.Has(id) {
			t := pk.types[n]
			if t.isTypevar {
				return args.Binds[n]
			} else if t.isGeneric() {
				return t.printedName(t.inst[args])
			} else {
				return n
			}
		}
		if pk.occConsts.Has(id) {
			v := pk.consts[n]
			return stringer.ToString(v)
		}
		if inCtor {
			if pk.occCtors.Has(id) {
				t := pk.ctors[n]
				instName := t.inst[args]
				return MangleCtorName(n, t.name(), instName)
			}
		}
		if pk.occPkgs.Has(id) {
			return pk.impRename[n]
		}
		return n
	}
}

func (td *TypeDesc) decl(instName string) []*ast.GenDecl {
	gd := &ast.GenDecl{}
	gd.Tok = token.TYPE
	gd.Specs = []ast.Spec{td.spec}
	if !td.isSingleton {
		return []*ast.GenDecl{gd}
	}

	vd := &ast.GenDecl{}
	vd.Tok = token.VAR
	vs := &ast.ValueSpec{
		Type:  &ast.Ident{Name: td.printedName(instName)},
		Names: []*ast.Ident{&ast.Ident{Name: instName}},
		/*
			Values: []ast.Expr{&ast.CompositeLit{
				Type: &ast.StructType{
					Fields: &ast.FieldList{},
				},
			}},
		*/
	}
	vd.Specs = []ast.Spec{vs}
	return []*ast.GenDecl{gd, vd}
}

func (pk *PkgDesc) print(wr *bufio.Writer, typedefs StrSet) {
	for _, tp := range pk.types {
		if !tp.isVisited {
			continue
		}
		if tp.isGeneric() {
			isFunc := tp.isSingleFunc()
			for typeArgs, instName := range tp.inst {
				p := newAstPrinter(wr, pk.renameFunc(typeArgs, false))
				if !typedefs.Has(instName) {
					// instName is printed once (this is how "merged" types work)
					if !isFunc {
						for _, d := range tp.decl(instName) {
							p.println(d)
						}
					}
					typedefs.Add(instName)
				}
				if len(tp.ctors) > 0 {
					p := newAstPrinter(wr, pk.renameFunc(typeArgs, true))
					for _, f := range tp.ctors {
						p.println(f)
					}
				}
				for _, f := range tp.methods {
					if isFunc {
						f.Recv = nil
						f.Name = &ast.Ident{Name: instName}
					}
					p.println(f)
				}
			}
		}
	}
}
