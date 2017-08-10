package main

// Foo is a useless const in the scheme of things
const Foo = "item"

// B is a mock type just for the sake of adding a type to const declarations
type B string

// Contents of this block shouldn't really be used for anything
const (
	Bar B = "bar"
	Baz B = "baz"
)
