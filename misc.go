package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

var symCounter int64 = 123

func GenSymbol(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, atomic.AddInt64(&symCounter, 1))
}

func DictStr(d map[string]string) (keystr, str string) {
	if len(d) == 0 {
		return "", ""
	}
	keys := make([]string, 0, len(d))
	for k, _ := range d {
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

// Immutable
type TypeArgs struct {
	Binds map[string]string // typevar -> replacement
	Key   string            // unique string for targsCache
	Shape string            // unique string of Binds map keys
}

var targsCache map[string]*TypeArgs = make(map[string]*TypeArgs)
var bindsCacheLock sync.Mutex

func TypeArgsOf(m map[string]string) *TypeArgs {
	if len(m) == 0 {
		return nil
	}
	bindsCacheLock.Lock()
	defer bindsCacheLock.Unlock()
	shape, key := DictStr(m)
	if b, ok := targsCache[key]; ok {
		return b
	} else {
		b := &TypeArgs{m, key, shape}
		targsCache[key] = b
		return b
	}
}

func (b *TypeArgs) Len() int {
	if b == nil {
		return 0
	}
	return len(b.Binds)
}

type localErr struct {
	error
}

// recoverTo recovers local panics
// Public funcs must use it to catch localPanic(): defer recoverTo(&err)
func recoverTo(pe *error) {
	r := recover()
	if r == nil {
		return
	}
	if e, ok := r.(localErr); ok {
		*pe = e
	} else {
		panic(r)
	}
}

// localPanic: Public func can't call this func without defer recoverTo()
func localPanic(e error) {
	if e != nil {
		panic(localErr{e})
	}
}

func localPanicf(format string, a ...interface{}) {
	localPanic(fmt.Errorf(format, a...))
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
	for _, dir := range filepath.SplitList(os.Getenv("GOPATH")) {
		fullPath := filepath.Join(dir, "src", pkg)
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
