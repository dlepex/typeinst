package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// orig - original ctor name, gen - generic type name, inst - instantiated type name
func MangleCtorName(orig, gen, inst string) string {
	n, isUpper := StrUpcase(orig)
	inst, _ = StrUpcase(inst)
	gen, _ = StrUpcase(gen)
	g := strings.Index(n, gen)

	if g < 0 {
		n = orig + inst
	} else {
		n = orig[0:g] + inst + n[g+len(gen):]
	}
	return StrEnsureCase(n, isUpper)
}

// "non-root generic types"
// orig - original dependant type name, gen - "parent" type name, inst - instantiated "parent" type name
func MangleDepTypeName(orig, gen, inst string) string {
	n, isUpper := StrUpcase(orig)
	gen, _ = StrUpcase(gen)
	inst, _ = StrUpcase(inst)
	g := strings.Index(n, gen)
	if g < 0 {
		n = inst + n
	} else {
		n = orig[0:g] + inst + n[g+len(gen):]
	}
	return StrEnsureCase(n, isUpper)
}

func StrUpcase(s string) (string, bool) {
	return StrReplaceFirst(s, unicode.IsUpper, unicode.ToUpper)
}
func StrLocase(s string) (string, bool) {
	return StrReplaceFirst(s, unicode.IsLower, unicode.ToLower)

}

func StrEnsureCase(s string, isUpper bool) (result string) {
	if isUpper {
		result, _ = StrUpcase(s)
	} else {
		result, _ = StrLocase(s)
	}
	return
}

func StrReplaceFirst(s string, isMapped func(rune) bool, doMap func(rune) rune) (string, bool) {
	if s == "" {
		return "", true
	}
	r, p := utf8.DecodeRuneInString(s)
	if isMapped(r) {
		return s, true
	} else {
		return string(doMap(r)) + s[p:], false
	}
}
