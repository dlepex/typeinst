// Le terme panique est une référence au dieu Pan en mythologie grecque
package pan

import (
	"fmt"
	"runtime"
	"strings"
)

// Bounded simplifies the implemenetation of "internal" panic i.e. panic that doesn't cross package boundaries
// The idea is that the public API of your package never leaks internal panics, see RecoverTo() & Panic() methods.
type Bounded struct {
	pkg string // todo:  caller pkg name
}

type errWrap struct {
	e error // wrapped error
	b *Bounded
}

// must be called to initialize private var of your package
func NewBounded() *Bounded {
	return &Bounded{callerPkg()}
}

// Public API of your package must use RecoverTo to translate "bounded panic" into error.
func (b *Bounded) RecoverTo(errPtr *error) {
	r := recover()
	if r == nil {
		return
	}
	if w, ok := r.(*errWrap); ok && w.b == b {
		// our own panic: catch.
		*errPtr = w.e
	} else {
		// alien panic: rethrow.
		panic(r)
	}
}

// Panic must never be called by public API of your pkg, without `defer b.RecoverTo(&err)`
func (b *Bounded) Panic(e error) {
	if e != nil {
		panic(&errWrap{e, b})
	}
}

func (b *Bounded) Panicf(format string, a ...interface{}) {
	b.Panic(fmt.Errorf(format, a...))
}

func callerPkg() string {
	pc, _, _, _ := runtime.Caller(2)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	pkg := ""
	if parts[pl-2][0] == '(' {
		pkg = strings.Join(parts[0:pl-2], ".")
	} else {
		pkg = strings.Join(parts[0:pl-1], ".")
	}
	return pkg
}

func (w *errWrap) Error() string {
	return fmt.Sprintf("Panic leak from pkg: %s, error: %v", w.b.pkg, w.e)
}
