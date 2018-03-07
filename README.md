# __Typeinst__
Typeinst is a tool to automate the creation of concrete types ("type instances") based generic/template types.

Typeinst uses the special fake struct declaration (DSL-struct, for brevity)  as the description of what types should be generated:
```go
import (
	"some-fictional-generic-package1/redblack"
	"some-fictional-generic-package2/std"
)
//go:generate typeinst
type _typeinst struct {
	IntTreeSet	func(K int, V struct{}) redblack.TreeMap
	FloatTreeMap	func(K float64, V float64) redblack.TreeMap
	StrSet		func(E string)  std.Set
} 
```
Each field of DSL-struct defines the single concrete type.

For each field *DSL-func* describes the substitution of type variables, which happens by name, and the result of DSL-func is the generic type where this substitution takes place.

## __Usage__

Typeinst is to be used with `go generate`, it has no command line options: it uses DSL-struct as its sole "option".

1. Install the tool first: `go install github.com/dlepex/typeinst`
1. Declare DSL-struct in some file of your package, together with go-generate comment, as in the example above.
	* The DSL-struct name must start with `_typeinst` prefix, it is strongly recommended to have one DSL-struct per package, and to declare it in a separate file.
1. Run `go generate` on your package.
1. The result is `<file>_ti.go`, where `<file>` is a name of the file where DSL-struct is declared. The file is generated in the same package and it contains ALL concrete types described by DSL-struct.

## __Features__
- __Selective type instantiation__: Typeinst will only create the required *generic types*, not the whole generic package at once. 
- [Constructor functions](#constructor-function) support
- [Type merging](#type-merging) support
- The implementer of a generic package doesn't need to use any special comments or magic imports. Generic package is a rather normal package, where some types (or type aliases) serve as type variables.

## __Terminology__

#### Type variable

Type variables (type-parameters of generic types) are declared within generic package, usually as empty interface:
```go
type A = interface{} // Alias is the preferred form.
type B interface{} // This form is possible too but it may be less convenient than the alias-based if you want to use a generic package directly i.e. w/o typeinst.

type C interface { // Non-empty interface type variables can be used as well.
	Less(C) bool     // Type variable C can only be substituted by types having `Less()` method.
}
```

__The names of DSL-func parameters define what types will serve as type variables__

As an implementor of a generic package you may __optionally__ use the special "typevar"-comment:
```go
type E = interface{} //typeinst: typevar
```
This comment only provides better error messages, in case a user of your generic package omits a type variable in DSL-func. 

Please note that, if the "typevar"-comment was used for one type variable - it MUST be used for all of them (in the same generic package, of course).

#### Generic type

Type is considered generic if one or more type variables are [reachable](#reachability) from it.

Generic type `G` consists of:
- type declaration 
- member functions, i.e. functions with receiver `G` or `*G`
- *constructor functions*


*Root generic types* are the types that are explicitly instantiated (i.e. the results of DSL-funcs)

*Non-root generic types* are *reachable* from root types and need to be "implicitly" instantiated (and implicitly named). For instance, root type AVLTree requires non-root type AVLTreeNode.

If you don't like the implicit ("mangled") names of non-root types, you can always name them on your own by making them root-types i.e. add their explicit instantiation to DSL-struct.

#### Constructor function

Constructor functions of generic type `G` is a  free-standing (no receiver) function that returns:
- `G`, `*G`, `[]G`, `[n]G`
- or any _two_ levels of those: `**G`, `*[42]G`, but not: `[]*[]G`.
- in case of multiple return values: only the first return var is checked.


Constructor functions usually have names started with `New`, but this is not enforced.

#### Reachability

Type B is directly reachable from type A if it occurs in:
- type A declaration
- signatures of type A member or constructor functions (bodies are not scanned!)

Reachability is a transitive, non-symmetric relation.

#### Generic package

*Generic package* contains *generic types* and their *type variables*. 

Generic package may import other packages. However, imported packages are never treated as "generic" themselves.

__Avoid non-generic code in generic package__, move it to separate non-generic package if needed.

Non-generic code includes:
- free-standing functions that are not constructors of generic types
- non-generic types
- vars and consts decl

#### Type merging

Type merging allows assembling a concrete type from multiple orthogonal behavioral parts. 

Partial types must be structurally the same i.e. have the same type expr after substitution. Note that typeinst does not check this property, your generated code simply will not compile if it is violated.

```go
type T = interface{} 
type SliceF T[] // this type implements filtering "methods"
type SliceA T[] // and this - aggregation "methods" (it may be declared in another generic package)
```

The merged type `IntSlice`  based on these two behavioral sub-units may be created using multiple return types in DSL-func:

```go
//go:generate typeinst
type _typeinst struct {
	IntSlice	func(T int) (somepkg.SliceA, somepkg.SliceF)
} 
```
`IntSlice` will contain both filtering and aggregation methods.

## __Limitations__

1. Imports in generic packages:
	- must have consistent import names across all files of the same generic package 
	- dot `.` import is not allowed
2. Type variables cannot be substituted by:
	- "anonymous" non-empty struct [solution: use named types or type alias]
	- "anonymous" non-empty interface [solution: the same]
3. Typeinst is type-based and so free standing functions (except constructors) cannot be generic. [solution: use functions with receiver instead, ultimately empty(`struct{}`) named types can be used as dummy receivers].
4. [Read generic package section](#generic-package)


## __Implementation notes__

- AST rewriting is not used. "Rewriting" happens simultaneously with printing AST to file. For that purpose, the standard "go/printer" package was slightly modified: "RenameFunc" was added to Config, it has 1 call site. 
- Typeisnt has been used to generate its own piece: see "gentypes.go" file.

## todo

- [ ] alternative dsl form: func(tv1-type, tv1-value, tv2-type, tv2-value,...)
- [ ] more tests & travis cfg




