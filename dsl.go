package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"reflect"
	"strings"
)

type (
	DSL struct {
		Imports
		Items   []*DSLItem
		PkgName string
	}

	PkgTypePair struct {
		PkgName string
		Type    string
	}

	DSLItem struct {
		InstName     string
		GenericTypes []PkgTypePair
		TypeArgs     map[string]string
	}
)

const DefaultStructName = "_typeinst"

func ParseDSL(filename, structName string) (dsl *DSL, err error) {
	defer recoverTo(&err)
	if structName == "" {
		structName = DefaultStructName
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	dsl = &DSL{
		PkgName: f.Name.Name,
	}
	imports := Imports{}

	for _, spec := range f.Imports {
		if err := imports.AddSpec(spec); err != nil {
			return nil, dslErrorf("bad imports: %v", err)
		}
	}

	stringer := astStringer{}

	parseFunc := func(it *DSLItem, t *ast.FuncType) {
		typeVarsPkgs := NewStrSet()
		walker := pkgNameWalker(typeVarsPkgs)
		for _, field := range t.Params.List {
			typeVar := fieldName(field)
			if typeVar == "" {
				localErr{dslErrorf("typevar param in func requires name")}.panic()
			}
			ast.Walk(walker, field.Type)
			it.TypeArgs[typeVar] = stringer.ToString(field.Type)
		}

		for pkgname, _ := range typeVarsPkgs {
			dsl.Imports.Add(pkgname, imports.RequireNamed(pkgname))
		}

		qtset := NewStrSet()
		for _, field := range t.Results.List {
			if len(field.Names) > 0 {
				localErr{dslErrorf("function result cannot have field names")}.panic()
			}
			pair := parseGenericTypeExpr(field.Type)
			if pair.PkgName == "" {
				localErr{dslErrorf("generic type cannot be local - it must be imported from another package")}.panic()
			}
			qt := pair.QualifiedType()
			if qtset.Has(qt) {
				localErr{dslErrorf("repeated generic type: %v", qt)}.panic()
			}
			qtset.Add(qt)
			pair.PkgName = imports.RequireNamed(pair.PkgName)
			it.GenericTypes = append(it.GenericTypes, pair)
		}
	}

	parseStruct := func(ts *ast.TypeSpec) {
		expr, ok := ts.Type.(*ast.StructType)
		if !ok {
			localErr{dslErrorf("struct type expected")}.panic()
		}
		if expr.Fields == nil || len(expr.Fields.List) == 0 {
			localErr{dslErrorf("empty struct")}.panic()
		}
		for _, field := range expr.Fields.List {
			it := &DSLItem{
				InstName: fieldName(field),
				TypeArgs: make(map[string]string),
			}
			ft, ok := field.Type.(*ast.FuncType)
			if !ok {
				localErr{dslErrorf("struct fields must have func types, e.g: `func(K int, V string) MyMap`, found: field: %s type: %v ",
					it.InstName, reflect.TypeOf(ts.Type))}.panic()
			}
			parseFunc(it, ft)
			dsl.Items = append(dsl.Items, it)
		}
	}
	for _, decl := range f.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			if decl.Tok == token.TYPE {
				for _, spec := range decl.Specs {
					ts := spec.(*ast.TypeSpec)
					name := ts.Name.Name
					if strings.HasPrefix(name, structName) {
						parseStruct(ts)
						return
					}
				}
			}
		}
	}
	return nil, dslErrorf("delaration of dsl struct not found: %s", structName)
}

func fieldName(field *ast.Field) string {
	if len(field.Names) != 1 {
		localErr{dslErrorf("field must have one name in struct fields and func params/returns: %v", field.Names)}.panic()
	}
	return field.Names[0].Name
}

func dslErrorf(format string, args ...interface{}) error {
	return fmt.Errorf("dsl: "+format, args...)
}

type astStringer struct {
	printer.Config
	buf  *bytes.Buffer
	fset *token.FileSet
}

func (s *astStringer) ToString(node interface{}) string {
	if s.buf == nil {
		s.buf = bytes.NewBuffer(make([]byte, 64))
		s.fset = token.NewFileSet()
		s.Config.Mode = printer.RawFormat
	}
	s.buf.Reset()
	s.Fprint(s.buf, s.fset, node)
	return s.buf.String()
}

type pkgNameWalker map[string]struct{}

func (w pkgNameWalker) Visit(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.SelectorExpr:
		switch n := n.X.(type) {
		case *ast.Ident:
			StrSet(w).Add(n.Name)
		}
	}
	return w
}

func parseGenericTypeExpr(t ast.Expr) PkgTypePair {
	switch t := t.(type) {
	case *ast.Ident:
		return PkgTypePair{"", t.Name}
	case *ast.SelectorExpr:
		typ := t.Sel.Name
		switch t := t.X.(type) {
		case *ast.Ident:
			return PkgTypePair{t.Name, typ}
		}
	}
	localErr{dslErrorf("unexpected type expr for generic type: %v in expr: %v", reflect.TypeOf(t), t)}.panic()
	return PkgTypePair{}
}

func (p PkgTypePair) QualifiedType() string {
	if p.PkgName == "" {
		return p.Type
	}
	return p.PkgName + "." + p.Type
}
