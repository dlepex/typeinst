# __Typeinst__
Typeinst is a tool to automate the creation of concrete types ("type instances") from generic/"template" types.

Typeinst uses the special fake struct declaration (DSL-struct, for brevity)  as the description of what types should be generated:
```go
import (
  "github.com/dlepex/genericlib/set"
	"github.com/dlepex/genericlib/slice"
	"some/thirparty/generic/package/redblack"
)
//go:generate typeinst
type _typeinst struct {
  StrSet		func(E string)  set.Set
  floats    func(E float64) slice.Ops
	IntTreeSet	func(K int, V struct{}) redblack.TreeMap
	FloatTreeMap	func(K float64, V float64) redblack.TreeMap

}
```
Each field of DSL-struct defines the single concrete type.

For each field *DSL-func* describes the substitution of type variables and the result of DSL-func is the generic type where this substitution takes place. The substitution is done by name.

## __Usage__

Typeinst is to be used with `go generate`, it has no command line options: it uses DSL-struct as its sole "option".

1. Install the tool first: `go get github.com/dlepex/typeinst`
1. Declare DSL-struct in some file of your package, together with go-generate comment, as in the example above.
	* The DSL-struct name must start with `_typeinst` prefix, it is strongly recommended to have one DSL-struct per package, and to declare it in a separate file.
1. Run `go generate` on your package.
1. The result is `<file>_ti.go`, where `<file>` is a name of the file where DSL-struct is declared. The file is generated in the same package and it contains ALL concrete types described by DSL-struct.
1. This repo https://github.com/dlepex/genericlib contains some usefull generic types e.g. generic slice ops and generic set


## __Features__
- __Selective type instantiation__: Typeinst only generates the requested types, not the whole generic package at once: this tool is type-based, not package-based.
- [Constructor functions](#constructor-function) support
- [Type merging](#type-merging) support
- No mandatory magic comments, and no magic imports in generic code
- Special support for [empty singleton generic types](#empty-singleton-generic-types).

## __Terminology__

### __Type variable__

Type variables (type-parameters of generic types) are declared within generic package, usually as an empty interface:
```go
type E interface{}
type A = interface{} // Alias form is ok too

type C interface { // Non-empty interface type variables can be used as well.
	Less(C) bool     // Type variable C can only be substituted by types having `Less()` method.
}
```
__The names of DSL-func parameters define what types will serve as type variables__

As an implementor of a generic package you may __optionally__ use the special "typevar"-comment:
```go
type E = interface{} //typeinst: typevar
```
This comment provides error message, in case a user of your generic package omits the type variable in DSL-func.

Please note that, if the "typevar"-comment was used for one type variable, it MUST be used for the rest of them (in the same generic package).

### __Generic type__

A type is considered generic if it [depends on](#type-dependency-relation) at least one type variable.

Generic type `G` consists of:
- type declaration
- methods (functions with receiver `G` or `*G`)
- [constructor functions](#constructor-function)


*Root generic types* are the types that are explicitly instantiated (i.e. the results of DSL-funcs)

Root types may [depend on](#type-dependency-relation) *non-root generic types*, non-root types are implicitly instantiated and implicitly named.
For instance, hypothetical root type `AVLTree` depends on non-root type `AVLTreeNode`.

If you do not like the implicit ("mangled") names of non-root types, you can always name them on your own by making them root, i.e. by adding their explicit instantiation to DSL-struct.

### __Constructor function__

Constructor function of generic type `G` is a function that returns:
- `G`, `*G`, `[]G`, `[n]G`
- or their combination (e.g. `**G`, `*[42]G`,`[][][]*G`), max nesting depth is 16.
- in case of multiple return values: only the first return var is checked.


Constructor functions usually have names started with `New`, but this is not enforced.

### __Type dependency relation__

Type A directly _depends on_ type B if type B occurs in:
- type A declaration
- signatures of type A methods or constructor functions

Type dependency is a transitive, non-symmetric relation.

### __Generic package__

*Generic package* contains [generic types](#generic-type) and their [type variables](#type-variable)

Generic packages **cannot contain non-generic code**, move it to separate non-generic package if needed.

Non-generic code includes:
- functions (w/o receiver), excluding constructors of generic types
- non-generic types and their methods
- var declarations

Const declarations are allowed in generic packages. Typeinst directly substitutes constants by their values.

Generic package may import other packages. Imported packages are never treated as generic themselves, i.e. a generic type from one package cannot depend on a generic type from another package.

### __Type merging__

Type merging allows an instantiated type to be assembled from multiple orthogonal behavioral parts (or in other words: non-intersecting method sets).

"Behavioral parts" must have the same declared type after the substitution of type variables.

```go
type T = interface{}
type SliceF T[] // this type has filtering methods
func (a SliceF) Filter(...) ... {...}

type SliceA T[] // and this - aggregation methods: it may be declared in another generic package, with another (differently named) type variable.
func (a SliceA) Reduce(...) ... {...}
```

The merged type `IntSlice` based on this 2 types may be created using multiple return types in DSL-func:

```go
//go:generate typeinst
type _typeinst struct {
	IntSlice	func(T int) (somepkg.SliceA, somepkg.SliceF)
}
```
`IntSlice` will contain both filtering and aggregation methods.

### __Empty singleton generic types__

ESGT are declared as empty structs and serve as dummy receivers for their methods, and thus
they can be used for generic function imitation. Typeinst is type-based and it is impossible to create generic functions directly.

Here is an example:
```go
type E interface{} // E is a type var
type ChanMerge struct{} // ESGT which depends on E through its method Merge i.e. this type is generic

func (_ ChanMerge) Merge(cs ...<-chan E) <-chan E {...}
```
**For ESGT Typeinst will not only generate a type-declaration, but also a var-declaration**

```go
//go:generate typeinst
type _typeinst struct {
	intChannels	func(E int) (somepkg.ChanMerge)
}
```
In this example Typeinst will generate _intChannels var_ and  _intChannelsType type_.

As a side note, since ESGT are just named empty structs, they are potentially [type-mergeable](#type-merging).

## __Limitations__

1. Imports in generic packages:
	- must have consistent import names across all files of the same generic package
	- dot `.` import is not allowed
2. Type variables cannot be substituted by:
	- "anonymous" non-empty struct [solution: use named types or type alias]
	- "anonymous" non-empty interface [solution: the same]
3. Functions w/o receiver (except constructors) cannot be generic [solution: [ESGT](#empty-singleton-generic-types)]
4. [Read generic package section](#generic-package)
5. Not all errors are checked during code generation, some of them will potentially result in uncompilable code:
	- non-generic code in generic package
	- merging unmergeable types
	- identifier name clashes or shadowing
1. It is worth to remember that Typeinst is a code generator and not a typechecker, and that in many cases `interface{}` is ok.

## __Implementation notes__

- AST rewriting is not used. Identifier substitution happens simultaneously with printing AST to file. For that purpose, the standard "go/printer" package was slightly modified: extra field `RenameFunc` was added to the `Config` struct.
- Typeinst has been used to generate a part of itself: [gentypes.go](https://github.com/dlepex/typeinst/blob/master/gentypes.go)
