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
	fatalIfErr(Run(gofile))
}

// Run - runs tool on a file
// added for more convenient testing
func Run(gofile string) (err error) {
	defer bpan.RecoverTo(&err)
	implFile := implFilename(gofile, fileSuffix)
	dsl, err := ParseDSL(gofile, "")
	bpan.Check(err)
	impl := newImpl(implFile, dsl.PkgName)
	dsl2Impl(dsl, impl)
	return
}

func dsl2Impl(dsl *DSL, impl *Impl) {
	for _, it := range dsl.Items {
		for _, g := range it.GenericTypes {
			p, err := impl.Package(g.PkgName, dsl.Imports)
			bpan.Check(err)
			log.Printf("dsl: type %s = %s with args: %v", it.InstName, g.Type, it.TypeArgs)
			bpan.Check(p.Inst(g.Type, it.InstName, it.TypeArgs))
		}
	}
	for path, pdesc := range impl.pkg {
		log.Printf("walk: %s", path)
		bpan.Check(pdesc.resolveGeneric())
	}
	log.Printf("printing...")
	bpan.Check(impl.Print())
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
