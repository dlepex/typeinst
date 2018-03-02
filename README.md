# __Typeinst__
Typeinst is a tool (codegen) to create concrete types based on generic types.  
Typeinst uses a special "magic" struct declaration as the description of what types should be created (*"instantiated"*):
```go
import (
	"some-generic-package1/maps"
	"some-generic-package2/sets"
)
//go:generate typeinst
type _typeinst struct { //
	IntOrdSet	func(K int, V struct{}) maps.OrderedMap
	FloatOrdMap	func(K float64, V float64) maps.OrderedMap
	StrSet		func(E int)  sets.Set
}
```
Since this magic struct defines the mini-language of new types creation, it will be called DSL-struct.

Each field of DSL-struct describes *instantiation*  of a single concrete type based on some generic type, 
the func ("DSL-func") describes substitution of type variables (which happens by name), and the result of DSL-func is generic type itself.

For example, lets take a look at the first field of this struct. It says: 
instantiate (create) new type *IntOrdSet* based on the code of generic type  *maps.OrderedMap* with the following indentifier substitution:
- replace *type variables* - *K* with type *int* and *V* with type *struct{}* 
- replace the type identifier *IntOrdSet* with *IntOrdMap*

### __Usage__

Typeinst is to be used with `go generate`, it has no arguments and uses DSL-struct as its sole "argument".

1. Install the tool first `go install github.com/dlepex/typeinst`
1. Declare DSL-struct in some file of your package, together with go-generate comment, as in example above.
	* The DSL-struct name must be started with `_typeinst` prefix, it's strongly recommended to have one DSL-struct per package, and to declare it in a separate file.
1. Run `go generate` on your package.
1. The result is `<file>_ti.go`, where `<file>` is a name of the file where DSL-struct is declared. The file is generated in the same package and it containes all instantiated types described by DSL-struct.

### __Features__

- Selective type instantiation: typeinst will only create the requested *generic types* of a *generic package*, not the whole package at once. 
- *Constructor functions* support
- *Type merging* support
- The implementer of a generic package doesn't need to use any special comments or imports. Generic package is a normal package, where some types (or aliases) serve as type variables.

### __Terminology__

#### Type variable

Type variables (parameters of generic types) are declared within generic package, usually as empty interface:
```go
type A = interface{} // alias is the preferred form.
type B interface{} // this form is possible too but it may be less convenient than alias for cases where you want to use generic package directly.

type C interface { // non-empty interface type variables are possible
	Less(C) bool     // type variable C can only be replaced by types having "Less" method.
}
```

__The names of DSL-func parameters define what types will serve as type variables__

As an implementor of a generic package you may use this magic comment:
```go
type E = interface{} //typeinst: typevar
```
This comment is optional and only provides better error messages, in case the user of your package omits this type variable in a DSL-func. However note: if this comment was used for one type variable - it MUST be used for all of them (in this generic package)

#### Generic type

Type is considered generic if one or more type variables are *reachable* from it.

Generic type `G` consists of:
- type declaration 
- memeber functions, i.e. functions with receiver `G` or `*G`
- *constructor functions*


*Root generic types* are the types that are explicitly instantiated (i.e. the results of DSL-funcs)

*Non-root generic types* are *reachable* from root types and need to be "implicitly" instantiated (and implictily named). (For instance, root type AVLTree might require non-root type AVLTreeNode)

If you don't like the implicit names of non-root types, you can always name them on your own by making them root.

#### Constructor function

Constructor functions of generic type G is any function that return:
- `G`, `*G`, `[]G`, `[n]G`
- or any _two_ levels of those: `**G`, `*[42]G`, but not: `[]*[]G`.
- in case of multiple return values: only the first return var is checked.


Constructor functions usually have names started with `New`, but this is not enforced.

#### Reachability

Type B is directly reachable from type A if it occurs in:
- type A declaration
- signatures of type A member or constructor functions (bodies are not scanned!)

Reachability a transitive, non-symmetric relation.

#### Generic package

Generic package contains *generic types* and their *type variables*.

Avoid non-generic code in generic package, move it to separate non-generic package if needed, and import this new pkg to generic.

Non-generic code includes:
- free standing functions: neither member, nor constructor of generic types
- non-generic types

Generic packages are not "recursive", i.e. generic package can't use generic types from other generic packages. But it's always possible to have a big fat generic package that containes all inter-dependant generic types (due to selective type instantiation feature).

#### Type merging


### Implementation notes

- AST rewriting is not used. Identifier "renaming" happens while printing. For that purpose, the standard "go/printer" package was slightly modified: "RenameFunc" was added to Config, it has 1 call site. 
- Fun fact: typeisnt has been used to generate its own piece: see "gentypes.go" file.





