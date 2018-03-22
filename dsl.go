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
	// DSL - dsl-struct description
	DSL struct {
		Imports
		Items   []*DSLItem
		PkgName string
	}
	// DSLItem corresponds to the field of dsl-struct
	DSLItem struct {
		InstName     string
		GenericTypes []PkgTypePair
		TypeArgs     map[string]string
	}
	// PkgTypePair tuple of type and its package
	PkgTypePair struct {
		PkgName string
		Type    string
	}
)

const defaultStructName = "_typeinst"

// ParseDSL parses and rertrieves dsl-struct
func ParseDSL(filename, structName string) (dsl *DSL, err error) {
	defer bpan.RecoverTo(&err)
	if structName == "" {
		structName = defaultStructName
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
			return nil, fmt.Errorf("bad imports: %v", err)
		}
	}

	stringer := astStringer{}

	parseFunc := func(it *DSLItem, t *ast.FuncType) {

		typeVarsPkgs := NewStrSet()
		walker := pkgNameWalker(typeVarsPkgs)
		estr := func(s string) string {
			return s + " [in dsl-struct field: " + it.InstName + "]"
		}
		if t.Params == nil || len(t.Params.List) == 0 {
			bpan.Panicf(estr("dsl-func has no arguments i.e. typevar substitutions"))
		}
		if t.Results == nil || len(t.Results.List) == 0 {
			bpan.Panicf(estr("dsl-func has no result i.e. generic type"))
		}
		for _, field := range t.Params.List {
			typeVar := fieldName(field)
			if typeVar == "" {
				bpan.Panicf(estr("typevar param in func requires name"))
			}
			ast.Walk(walker, field.Type)
			it.TypeArgs[typeVar] = stringer.ToString(field.Type)
		}

		for pkgname := range typeVarsPkgs {
			bpan.Check(dsl.Imports.Add(pkgname, imports.requireNamed(pkgname)))
		}

		qtset := NewStrSet()
		for _, field := range t.Results.List {
			if len(field.Names) > 0 {
				bpan.Panicf(estr("dsl-func result cannot have field names"))
			}
			pair := parseGenericTypeExpr(field.Type)
			if pair.PkgName == "" {
				bpan.Panicf(estr("generic type cannot be local, it must be imported from another package"))
			}
			qt := pair.qualifiedType()
			if qtset.Has(qt) {
				bpan.Panicf(estr("merging repeated generic type: %v"), qt)
			}
			qtset.Add(qt)
			pair.PkgName = imports.requireNamed(pair.PkgName)
			it.GenericTypes = append(it.GenericTypes, pair)
		}
	}

	parseStruct := func(ts *ast.TypeSpec) {
		expr, ok := ts.Type.(*ast.StructType)
		if !ok {
			bpan.Panicf("struct type expected")
		}
		if expr.Fields == nil || len(expr.Fields.List) == 0 {
			bpan.Panicf("empty struct")
		}
		for _, field := range expr.Fields.List {
			it := &DSLItem{
				InstName: fieldName(field),
				TypeArgs: make(map[string]string),
			}
			ft, ok := field.Type.(*ast.FuncType)
			if !ok {
				bpan.Panicf("struct fields must have func types, e.g: `func(K int, V string) MyMap`, found: field: %s type: %v ",
					it.InstName, reflect.TypeOf(ts.Type))
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
	return nil, fmt.Errorf("delaration of dsl struct not found: %s", structName)
}

func fieldName(field *ast.Field) string {
	if len(field.Names) != 1 {
		bpan.Panicf("field must have one name in struct fields and func params/returns: %v", field.Names)
	}
	return field.Names[0].Name
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
	_ = s.Fprint(s.buf, s.fset, node)
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
	bpan.Panicf("unexpected type expr for generic type: %v in expr: %v", reflect.TypeOf(t), t)
	return PkgTypePair{}
}

func (p PkgTypePair) qualifiedType() string {
	if p.PkgName == "" {
		return p.Type
	}
	return p.PkgName + "." + p.Type
}
