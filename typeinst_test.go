package main

import (
	"os/exec"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsage(t *testing.T) {
	e := run("testdata/usage/case1.go")
	assert.NoError(t, e)
	cmd := exec.Command("go", "build", "github.com/dlepex/typeinst/testdata/usage")
	b, e := cmd.CombinedOutput()
	if e != nil {
		t.Errorf(string(b))
	}
}

func run(f string) error {
	p := packagePath(path.Join("github.com/dlepex/typeinst", f))
	return Run(p)
}
