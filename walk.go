package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

type (
	// Impl - root structure, it aggregates all generic packages - it will be printed as a single implementation file
	Impl struct {
		pkg        map[string]*PkgDesc
		imports    Imports
		instNames  StrSet
		pkgName    string // generated package name
		outputFile string // generated file name
	}

	// Generic package desc - all types and their functions
	PkgDesc struct {
		name      string
		types     map[string]*TypeDesc     // all package types by name
		ctors     map[string]*TypeDesc     // ctor name to type (it belongs)
		typevars  StrSet                   // set of types (their names) that used as type variables
		generic   StrSet                   // set of generic types
		funcs     map[string]*ast.FuncDecl // free standing funcs (i.e. no recever), excluding type ctors
		impRename map[string]string        // what packages should be renamed within pkg AST: name -> newname
		strict    bool                     // strict mode means all typevars of the pkg are markerd with special comment "//typeinst: typevar"
		occTypes  AstIdentSet              // occurences of types identifiers in AST (that may be renamed)
		occPkgs   AstIdentSet              // occurences of packages identifiers in AST (that must be renamed)
		occCtors  AstIdentSet              // occurences of constructor functions ...
	}

	TypeDesc struct {
		spec     *ast.TypeSpec
		methods  []*ast.FuncDecl
		ctors    []*ast.FuncDecl      // ctor is func w/o receiver, returning instances of: G,[]G,*G,**G,...(any two levels of * or []) where G is generic type (root or "dependant")
		inst     map[*TypeArgs]string // typeargs -> instname; (map nonempty only for generic types)
		typevars StrSet               // set is populated typevars upon which this generic type depends
		typevar  bool                 // does this type serves as a typevar?
		visited  bool                 // was this type ever visited from any "root" generic type (root is what passed to Inst() func)
	}
)

func NewImpl(outputFile, pkgName string) *Impl {
	return &Impl{
		pkg:        make(map[string]*PkgDesc),
		imports:    Imports{},
		instNames:  NewStrSet(),
		outputFile: outputFile,
		pkgName:    pkgName,
	}
}

func (td *TypeDesc) String() string {
	return fmt.Sprintf("{fn: %v, cc: %v, t: %v}", td.methods, td.ctors, td.typevar)
}

func (td *TypeDesc) name() string {
	return td.spec.Name.Name
}

func (td *TypeDesc) addFunc(f *ast.FuncDecl) {
	if td.typevar {
		bpan.Panicf("Typevar %s can't be func receiver: %s", td.name(), f.Name.Name)
	}
	td.methods = append(td.methods, f)
}

func (td *TypeDesc) addCtor(f *ast.FuncDecl) {
	if td.typevar {
		bpan.Panicf("Typevar %s can't have constructors: %s", td.name(), f.Name.Name)
	}
	td.ctors = append(td.ctors, f)
}

func (td *TypeDesc) initBinds() {
	if td.inst == nil {
		td.inst = make(map[*TypeArgs]string)
	}
}

func (td *TypeDesc) canBeTypevar() bool {
	return len(td.ctors) == 0 && len(td.methods) == 0
}

func (t *TypeDesc) IsGeneric() bool { return len(t.typevars) != 0 }

func (t *TypeDesc) inheritFrom(parent *TypeDesc) {
	if t == parent {
		return
	}
	t.initBinds()
	for b, instName := range parent.inst {
		if _, has := t.inst[b]; !has {
			t.inst[b] = MangleDepTypeName(t.name(), parent.name(), instName)
		}
	}
}

type tdescDict map[string]*TypeDesc

func (m tdescDict) get(name string) *TypeDesc {
	if t, ok := m[name]; ok {
		return t
	} else {
		t = &TypeDesc{}
		m[name] = t
		return t
	}
}

func (impl *Impl) Package(pkgPath string, imports Imports) (pkg *PkgDesc, err error) {
	if p, ok := impl.pkg[pkgPath]; ok {
		return p, nil
	}
	defer bpan.RecoverTo(&err)
	types := tdescDict(make(map[string]*TypeDesc))
	funcs := make(map[string]*ast.FuncDecl)
	tpvars := NewStrSet()
	fset := token.NewFileSet()
	pkgpath := packagePath(unquote(pkgPath))
	if pkgpath == "" {
		return nil, fmt.Errorf("no such package: %s", pkgPath)
	}
	m, err := parser.ParseDir(fset, pkgpath, pkgFileFilter, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	for _, pkg := range m {
		for fn, f := range pkg.Files {
			for _, spec := range f.Imports {
				if err := imports.AddSpec(spec); err != nil {
					return nil, fmt.Errorf("bad imports(...) in package: %s, file: %s, %v", pkgpath, fn, err)
				}
			}
			for _, decl := range f.Decls {
				switch decl := decl.(type) {
				case *ast.FuncDecl:
					if r := receiverType(decl); r != "" {
						tdef := types.get(r)
						tdef.addFunc(decl)
					} else {
						funcs[decl.Name.Name] = decl
					}
				case *ast.GenDecl:
					if decl.Tok == token.TYPE {
						for _, spec := range decl.Specs {
							tsp := spec.(*ast.TypeSpec)
							name := tsp.Name.Name
							tdef := types.get(name)
							tdef.spec = tsp
							if tsp.Comment != nil {
								for _, c := range tsp.Comment.List {
									if tdef.ParseSpecialComment(c.Text) {
										if tdef.typevar {
											tpvars[name] = struct{}{}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	var impRename map[string]string
	if impl.imports.IsEmpty() {
		impl.imports = imports
	} else {
		impRename = impl.imports.Merge(imports)
	}

	pkg = &PkgDesc{pkgPath, types, make(map[string]*TypeDesc), tpvars, NewStrSet(), funcs, impRename, len(tpvars) > 0,
		NewAstIdentSet(), NewAstIdentSet(), NewAstIdentSet()}
	pkg.detectCtors()
	impl.pkg[pkgPath] = pkg
	return
}

func unpackRecur(depth uint32, t ast.Node) string {
	if depth == 0 {
		return ""
	}
	switch t := t.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return unpackRecur(depth-1, t.X)
	case *ast.ArrayType:
		return unpackRecur(depth-1, t.Elt)
	default:
		return ""
	}
}

func unpackCtorRet(fd *ast.FuncDecl) string {
	r := fd.Type.Results
	if r != nil && len(r.List) > 0 {
		return unpackRecur(16, r.List[0].Type)
	}
	return ""
}

func pkgFileFilter(info os.FileInfo) bool {
	name := info.Name()
	if strings.HasSuffix(name, "_test.go") || !strings.HasSuffix(name, ".go") {
		return false
	}
	return true
}

func receiverType(fd *ast.FuncDecl) string {
	if fd.Recv == nil {
		return ""
	}
	t := fd.Recv.List[0].Type
	var name string
	switch t := t.(type) {
	case *ast.Ident:
		name = t.Name
	case *ast.StarExpr:
		x := t.X
		if id, ok := x.(*ast.Ident); ok {
			name = id.Name
		} else {
			log.Printf("Unsupported star(*) receiver type: %v", reflect.TypeOf(x))
			return ""
		}
	default:
		log.Printf("Unsupported receiver type: %v", reflect.TypeOf(t))
		return ""
	}
	return name
}

const commentPrefix string = "//typeinst:"

func (td *TypeDesc) ParseSpecialComment(text string) bool {
	if strings.HasPrefix(text, commentPrefix) {
		text = strings.TrimPrefix(text, commentPrefix)
		args := strings.Fields(text)
		if len(args) != 0 {
			verb := args[0]
			args = args[1:]
			switch verb {
			case "typevar":
				td.typevar = true
				return true
			default:
				log.Printf("ignoring illegal '%s'-comment unknown verb: %s", commentPrefix, verb)
			}
		} else {
			log.Printf("ignoring empty '%s'-comment", commentPrefix)
		}
	}
	return false
}

func (pd *PkgDesc) detectCtors() {
	ctors := []string{}
	for fname, fd := range pd.funcs {
		if r := unpackCtorRet(fd); r != "" {
			if tdef, ok := pd.types[r]; ok {
				tdef.addCtor(fd)
				pd.ctors[fname] = tdef
				pd.occCtors[fd.Name] = struct{}{}
				ctors = append(ctors, fname)
			}
		}
	}
	for _, fname := range ctors {
		delete(pd.funcs, fname)
	}
}

func (pk *PkgDesc) Inst(typName, instName string, typeArgs map[string]string) error {
	t, ok := pk.types[typName]
	if !ok {
		return fmt.Errorf("Type %s not found in package %s", typName, pk.name)
	}
	for tv, _ := range typeArgs {
		if !pk.typevars.Has(tv) {
			if pk.strict {
				return fmt.Errorf("strict mode: type %s cannot be a typevar in package  %s", tv, pk.name)
			}
			t, ok := pk.types[tv]
			if !ok {
				return fmt.Errorf("type %s (a typevar) not found in package %s", tv, pk.name)
			}
			if !t.canBeTypevar() {
				return fmt.Errorf("type %s cannot be a typevar in package  %s", tv, pk.name)
			}
			pk.typevars.Add(tv)
			t.typevar = true
		}
	}
	b := TypeArgsOf(typeArgs)
	if _, has := t.inst[b]; has {
		return fmt.Errorf("Type %s instantiated repeatedly with the same (type) arguments (%s) in package %s", typName, b.Key, pk.name)
	}
	if shape := t.shape(); shape != nil && shape.Shape != b.Shape {
		return fmt.Errorf("Type %s cannot be instantiated several times with inconsitent typevars (<%s> != <%s>) in package %s",
			typName, b.Shape, shape.Shape, pk.name)
	}
	t.initBinds()
	t.inst[b] = instName
	pk.generic.Add(typName)
	return nil
}

func fmtTypeName(parent, tn string) string {
	r, _ := utf8.DecodeRuneInString(tn)
	if unicode.IsUpper(r) {
		return parent + tn
	} else {
		return "privt_" + parent + tn
	}
}

func (pk *PkgDesc) resolveRecur(td, parent *TypeDesc, visited StrSet) {
	if visited.Has(td.name()) {
		return
	}
	td.visited = true
	visited.Add(td.name())
	if parent == nil {
		parent = td
	}
	depTypes := NewStrSet()
	td.typevars = NewStrSet()
	pk.walkType(td, func(params astWalkerParams) {
		id := params.id
		tn := id.Name
		if t, ok := pk.types[tn]; ok {
			if t != td {
				if t.typevar {
					td.typevars.Add(tn)
				} else {
					depTypes.Add(tn)
				}
			}
		}
	})
	for tn, _ := range depTypes {
		dept, _ := pk.types[tn]
		pk.resolveRecur(dept, parent, visited)
		if dept.IsGeneric() {
			pk.generic.Add(dept.name())
			// all typevars from dep type are inherited by "parent"
			for tn, _ := range dept.typevars {
				td.typevars.Add(tn)
			}
		}
	}
	if len(td.typevars) != 0 {
		td.inheritFrom(parent)
	} else {
		td.typevars = nil
	}
}

func (td *TypeDesc) shape() *TypeArgs {
	for any, _ := range td.inst {
		return any
	}
	return nil
}

func (pk *PkgDesc) ResolveGeneric() error {
	roots := make([]*TypeDesc, 0, len(pk.generic))
	for tn, _ := range pk.generic {
		t, _ := pk.types[tn]
		roots = append(roots, t)
	}
	for _, t := range roots {
		pk.resolveRecur(t, nil, NewStrSet())
	}

	for tn, _ := range pk.generic {
		gent, _ := pk.types[tn]
		b := gent.shape()
		for tv, _ := range gent.typevars {
			if _, has := b.Binds[tv]; !has {
				return fmt.Errorf("typevar %s is unbound for generic type %s in package %s", tv, tn, pk.name)
			}
		}
	}

	for _, t := range pk.types {
		if t.IsGeneric() {
			for ta, instName := range t.inst {
				log.Printf("resolved: type %s = %s with args: %v ", instName, t.name(), ta.Binds)
			}
			pk.walkTypeMarkOcc(t)
		}
	}
	return nil
}

func (pk *PkgDesc) markOccurences(p astWalkerParams) {
	n := p.id.Name

	if p.kind == ast.Typ || p.kind == ast.Bad {
		if pk.generic.Has(n) || pk.typevars.Has(n) {
			pk.occTypes.Add(p.id)
		}
	} else if p.kind == ast.Fun || p.kind == ast.Bad {
		if _, has := pk.ctors[n]; has {
			pk.occCtors.Add(p.id)
		}
	} else {
		if _, has := pk.impRename[n]; has {
			pk.occPkgs.Add(p.id)
		}
	}
}

func (pk *PkgDesc) walkType(t *TypeDesc, vf func(astWalkerParams)) {
	var reach astWalker = vf // "reachability" walker (it avoids body)

	ast.Walk(reach, t.spec.Type)

	for _, f := range t.methods {
		ast.Walk(reach, f.Type)
	}

	for _, f := range t.ctors {
		ast.Walk(reach, f.Type)
	}
}

func (pk *PkgDesc) walkTypeMarkOcc(t *TypeDesc) {
	var mark astWalker = pk.markOccurences
	ast.Walk(mark, t.spec.Type)
	pk.occTypes.Add(t.spec.Name)
	for _, f := range t.methods {
		ast.Walk(mark, f.Type)
		ast.Walk(mark, f.Body)
		ast.Walk(mark, f.Recv)
	}
	for _, f := range t.ctors {
		ast.Walk(mark, f.Type)
		ast.Walk(mark, f.Body)
	}
}

type astWalkerParams struct {
	id   *ast.Ident
	kind ast.ObjKind // Fun/Pkg/Typ
}

type astWalker func(params astWalkerParams)

func (w astWalker) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return w
	}
	switch node := node.(type) {
	case *ast.Ident:
		if node.Obj == nil {
			w(astWalkerParams{node, ast.Bad})
		} else {
			switch node.Obj.Kind {
			case ast.Typ, ast.Fun:
				w(astWalkerParams{node, node.Obj.Kind})
			}
		}
	case *ast.SelectorExpr:
		switch x := node.X.(type) {
		case *ast.Ident:
			if x.Obj == nil || x.Obj.Kind == ast.Pkg {
				w(astWalkerParams{x, ast.Pkg})
			}
		}
		return nil
	case *ast.ImportSpec:
		return nil
	}
	return w
}
