package main

import (
	"go/ast"
	"go/token"
	"log"
	"os"
	"path"
	"strings"
	"testing"

	prn "github.com/dlepex/typeinst/printer"
)

func TestPkgDef(t *testing.T) {
	impl := NewImpl("", "")
	impl.Package("github.com/dlepex/typeinst/testdata/test", Imports{})
	for k, p := range impl.pkg {
		t.Logf("%v # %v ## %s", k, p.typevars, p.types)
	}
}

func TestWriteFile(t *testing.T) {
	impl := NewImpl("", "")
	pk, err := impl.Package("github.com/dlepex/typeinst/testdata/g/maps", Imports{})
	ce(t, err)
	pk1, err := impl.Package("github.com/dlepex/typeinst/testdata/g/slices/filter", Imports{})
	ce(t, err)
	impl.outputFile = "/tmp/out.go"
	impl.pkgName = "testpkg"

	err = pk.Inst("Map", "Dict", map[string]string{
		"K": "int",
		"V": "[]struct{}",
	})
	ce(t, err)
	err = pk.Inst("Map", "DictOfFloat", map[string]string{
		"K": "int64",
		"V": "float32",
	})
	ce(t, err)

	err = pk1.Inst("Slice", "Ints", map[string]string{
		"T": "int",
	})
	ce(t, err)
	ce(t, pk.ResolveGeneric())
	ce(t, pk1.ResolveGeneric())

	ce(t, impl.Print())
}

func TestPkgInst(t *testing.T) {
	impl := NewImpl("", "")
	pd, err := impl.Package("github.com/dlepex/typeinst/testdata/test", Imports{})
	if err != nil {
		t.Error(err)
		return
	}
	err = pd.Inst("Slice", "Ints", map[string]string{
		"T": "int",
		"E": "string",
	})
	/*
		err = pd.Inst("Slice", "Floats", map[string]string{
			"T": "float",
			"E": "interface{}",
		})*/

	err = pd.Inst("Tux", "TuxOfFloat", map[string]string{
		"T": "float",
		"E": "interface{}",
	})

	if err != nil {
		t.Error(err)
		return
	}
	pd.impRename = make(map[string]string)
	pd.impRename["_un"] = "x"
	err = pd.ResolveGeneric()
	if err != nil {
		t.Error(err)
		return
	}
	conf := &prn.Config{}

	conf.Mode = prn.UseSpaces

	conf.RenameFunc = func(id *ast.Ident) string {
		if _, k := pd.occTypes[id]; k {
			return "<<" + id.Name + ">>"
		}
		if _, k := pd.occPkgs[id]; k {
			return "CODABRA"
		}
		return id.Name
	}

	t.Logf("IMPORTS %v", impl.imports.n2p)

	for tn, tp := range pd.types {
		if tp.visited {
			t.Logf("TYPE: %s <tv: %v> -->> binds: %v, gen: %v,  ctors %v", tn, tp.typevars, tp.inst, tp.IsGeneric(), tp.ctors)
			//ast.Print(nil, tp.spec.Type)
		}
		if len(tp.funcs) == 0 {
			continue
		}
		gd := ast.GenDecl{}
		gd.Tok = token.TYPE
		gd.Specs = []ast.Spec{tp.spec}
		fset := token.NewFileSet()
		conf.Fprint(os.Stdout, fset, &gd)
		os.Stdout.WriteString("\n")
		for _, f := range tp.ctors {
			conf.Fprint(os.Stdout, fset, f)
			os.Stdout.WriteString("\n")
		}
		for _, f := range tp.funcs {
			conf.Fprint(os.Stdout, fset, f)
			os.Stdout.WriteString("\n")
		}

	}
}

func Test2(t *testing.T) {
	a := strings.FieldsFunc("\"hello/world\"", func(r rune) bool {
		return r == '"' || r == '/'
	})
	t.Log(a[len(a)-1])
}

func TestPkg(t *testing.T) {
	im := Imports{}
	ce(t, im.Add("n", "p"))
	ce(t, im.Add("n1", "p1"))

	im1 := Imports{}

	ce(t, im1.Add("a", "p1"))
	ce(t, im1.Add("n1", "Q"))
	ce(t, im1.Add("a1", "Q1"))

	m := im.Merge(im1)
	t.Logf("rename %v result %v", m, im.n2p)
}

func ce(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("CheckErr: %v", err)
	}
}

func TestDsl(t *testing.T) {
	dsl, err := ParseDSL("/Users/dlepex/dev/go/work/src/github.com/dlepex/typeinst/testdata/usage/usg1.go", "")
	ce(t, err)

	for _, it := range dsl.Items {
		t.Logf("IT: %+v", it)
	}
	t.Logf("DSL: %v", dsl.n2p)
}

func Test21(t *testing.T) {
	p := "hello/абвгд.go"
	f := path.Base(p)
	pos := strings.LastIndex(f, ".go")
	if pos == -1 {
		log.Fatal("Not a .go file")
	}
	f = f[0:pos]
	f = path.Join(path.Dir(p), f+"_impl.go")
	t.Logf("[%s]", f)
}
