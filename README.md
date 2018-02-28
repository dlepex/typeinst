## typeinst

Typeinst generates code for parametrized/generic types based on DSL in a form of struct declaration. The declaration itself is a normal compilable Go code (which makes it possible to use IDE code completion). Also this keeps `go generate` comment really short!

DSL sample:
```go
import (
	"some-generic-package1/maps"
	"some-generic-package2/sets"
)
//go:generate typeinst
type _typeinst struct {
	IntOrdMap    func(K int, V int) maps.OrderedMap
  FloatOrdMap  func(K float64, V float64) maps.OrderedMap
	StrSet    	 func(E int)  sets.Set
}
```
*_typeinst* is special struct declaration, each field of which describes "instantiation" of a single concrete type based on some generic type. Here the word __instantiation__ has a meaning of code generation with all needed substitutions of identifiers.

Lets take the first field of this struct. It says: 
instantiate new type *IntOrdMap* based on the code of generic type  *maps.OrderedMap* with the following indentifier renaming:
- replace *type variables* *K* and *V* with type *int* 
- replace the type name *OrderedMap* with *IntOrdMap* 

### Features

- Selective type instantiation: typeinst will only instantiate the requested *generic types* of a *generic package*, not the whole package at once.
- Explicit result types naming (in the example above: IntOrdMap, FloatOrdMap, StrSet)
- *Constructor functions* support
- "Merged types" support
- The implementor of a generic package need not use any magic comments or magic imports. The only magic is DSL struct declaration.

#### __Terminology__

##### Type variable
Type variables are declared within generic packages, usually as empty interface alias:
```go
type E = interface{}
type E1 interface{} // this form is possible too, but may be less convenient for some cases

type E2 interface {
	Less(E2) bool
}
```

##### Generic type

##### Generic package

##### Type merging




### Implementation notes

- AST rewriting is not used, instead identifier renaming happens while printing. For that purpose, the modified version of "go/printer" package is used. The modification is very slight: the "RenameFunc" was added to Config. This function has single call site.
- Fun fact: typeisnt has been used to generate its own piece: see "gentypes.go" file.





