package main

import (
	"os"
	"os/exec"
	"path"
	"testing"
)

func TestUsage(t *testing.T) {
	run("testdata/usage/case1.go")
	cmd := exec.Command("go", "build", "github.com/dlepex/typeinst/testdata/usage")
	b, e := cmd.CombinedOutput()
	if e != nil {
		t.Errorf(string(b))
	}
}

func run(f string) {
	p := packagePath(path.Join("github.com/dlepex/typeinst", f))
	os.Setenv("GOFILE", p)
	main()
}
