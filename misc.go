package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/dlepex/typeinst/pan"
)

var bpan = pan.NewBounded()

var symCounter int64 = 9

func genSymbol(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, atomic.AddInt64(&symCounter, 1))
}

func dictStr(d map[string]string) (keystr, str string) {
	if len(d) == 0 {
		return "", ""
	}
	keys := make([]string, 0, len(d))
	for k := range d {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b := bytes.NewBuffer(make([]byte, 0, 64))
	bk := bytes.NewBuffer(make([]byte, 0, 64))
	for _, k := range keys {
		b.WriteString(k)
		b.WriteRune('=')
		b.WriteString(d[k])
		b.WriteRune(',')
		bk.WriteString(k)
		bk.WriteRune(',')
	}
	return bk.String(), b.String()
}

// TypeArgs is cached typevar bindings map, *TypeArgs is used as map key.
type TypeArgs struct {
	Binds map[string]string // typevar -> replacement
	Key   string            // unique string for targsCache
	Shape string            // unique string of Binds map keys
}

var targsCache = make(map[string]*TypeArgs)
var bindsCacheLock sync.Mutex

// TypeArgsOf returns cached result
func TypeArgsOf(m map[string]string) *TypeArgs {
	if len(m) == 0 {
		return nil
	}
	bindsCacheLock.Lock()
	defer bindsCacheLock.Unlock()
	shape, key := dictStr(m)
	if b, ok := targsCache[key]; ok {
		return b
	}
	b := &TypeArgs{m, key, shape}
	targsCache[key] = b
	return b
}

// Len returns number of bindings
func (b *TypeArgs) Len() int {
	if b == nil {
		return 0
	}
	return len(b.Binds)
}

func unquote(s string) string {
	q := "\""
	if strings.HasPrefix(s, q) {
		s = s[1:]
	}
	if strings.HasSuffix(s, q) {
		s = s[:len(s)-1]
	}
	return s
}

func packagePath(pkg string) string {
	cmd := exec.Command("go", "list", "-m", "all")
	b, err := cmd.CombinedOutput()
	if err != nil {
		//no go modules:
		return packagePathGopath(pkg, "src")
	}
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		sp := strings.Split(sc.Text(), " ")
		mod, ver := sp[0], ""
		if len(sp) == 2 {
			ver = sp[1]
		}
		if strings.HasPrefix(pkg, mod) {
			suffix := pkg[len(mod):]
			full := mod + "@" + ver + suffix
			return packagePathGopath(full, "pkg/mod")
		}
	}
	return ""
}

func packagePathGopath(pkg string, subdir string) string {
	for _, dir := range filepath.SplitList(os.Getenv("GOPATH")) {
		fullPath := filepath.Join(dir, subdir, pkg)
		if pathExists(fullPath) {
			return fullPath
		}
	}
	return ""
}

func pathExists(fpath string) bool {
	_, err := os.Stat(fpath)
	return !os.IsNotExist(err)
}
