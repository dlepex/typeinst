package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// MangleCtorName mangles constructor function name.
// orig - original ctor name, gen - generic type name, inst - instantiated type name
func MangleCtorName(orig, gen, inst string) string {
	n, isUpper := strUpcase(orig)
	inst, _ = strUpcase(inst)
	gen, _ = strUpcase(gen)
	g := strings.Index(n, gen)

	if g < 0 {
		n = orig + inst
	} else {
		n = orig[0:g] + inst + n[g+len(gen):]
	}
	return strEnsureCase(n, isUpper)
}

// MangleDepTypeName mangles non-root generic type name.
// orig - original dependant type name, gen - "parent" type name, inst - instantiated "parent" type name
func MangleDepTypeName(orig, gen, inst string) string {
	n, isUpper := strUpcase(orig)
	gen, _ = strUpcase(gen)
	inst, _ = strUpcase(inst)
	g := strings.Index(n, gen)
	if g < 0 {
		n = inst + n
	} else {
		n = orig[0:g] + inst + n[g+len(gen):]
	}
	return strEnsureCase(n, isUpper)
}

func strUpcase(s string) (string, bool) {
	return strReplaceFirst(s, unicode.IsUpper, unicode.ToUpper)
}
func strLocase(s string) (string, bool) {
	return strReplaceFirst(s, unicode.IsLower, unicode.ToLower)

}

func strEnsureCase(s string, isUpper bool) (result string) {
	if isUpper {
		result, _ = strUpcase(s)
	} else {
		result, _ = strLocase(s)
	}
	return
}

func strReplaceFirst(s string, isMapped func(rune) bool, doMap func(rune) rune) (string, bool) {
	if s == "" {
		return "", true
	}
	r, p := utf8.DecodeRuneInString(s)
	if isMapped(r) {
		return s, true
	}
	return string(doMap(r)) + s[p:], false
}
