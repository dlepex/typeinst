package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

//TODO DependentTypeName

// generated file suffix
const fileSuffix = "_ti"

func main() {
	gofile := os.Getenv("GOFILE")
	fmt.Printf("$GOPATH = %v\n$GOFILE = %v\n", os.Getenv("GOPATH"), gofile)

	implFile := implFilename(gofile, fileSuffix)

	dsl, err := ParseDSL(gofile, "")
	fatalIfErr(err)
	impl := NewImpl(implFile, dsl.PkgName)
	err = Dsl2Impl(dsl, impl)
	fatalIfErr(err)
}

func Dsl2Impl(dsl *DSL, impl *Impl) (err error) {
	defer bpan.RecoverTo(&err)
	check := bpan.Panic
	for _, it := range dsl.Items {
		for _, g := range it.GenericTypes {
			p, err := impl.Package(g.PkgName, dsl.Imports)
			check(err)
			log.Printf("dsl: type %s = %s with args: %v", it.InstName, g.Type, it.TypeArgs)
			check(p.Inst(g.Type, it.InstName, it.TypeArgs))
		}
	}
	for path, pdesc := range impl.pkg {
		log.Printf("walk: %s", path)
		check(pdesc.ResolveGeneric())
	}
	log.Printf("printing...")
	check(impl.Print())
	return
}

func implFilename(p, suf string) string {
	f := path.Base(p)
	pos := strings.LastIndex(f, ".go")
	if pos == -1 {
		log.Fatal("not a .go file")
	}
	f = f[0:pos]
	return path.Join(path.Dir(p), f+suf+".go")
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
