package fastjsonPackageInfo

var StructMap_3 map[string]string = map[string]string{
	"Parser": `// Parser parses JSON.
//
// Parser may be re-used for subsequent parsing.
//
// Parser cannot be used from concurrent goroutines.
// Use per-goroutine parsers or ParserPool instead.
type Parser struct {
	// b contains working copy of the string to be parsed.
	b []byte

	// c is a cache for json values.
	c cache
}`,
	"cache": `type cache struct {
	vs []Value
}`,
	"kv": `type kv struct {
	k string
	v *Value
}`,
	"Object": `// Object represents JSON object.
//
// Object cannot be used from concurrent goroutines.
// Use per-goroutine parsers or ParserPool instead.
type Object struct {
	kvs           []kv
	keysUnescaped bool
},type kv struct {
	k string
	v *Value
}`,
	"Value": `// Value represents any JSON value.
//
// Call Type in order to determine the actual type of the JSON value.
//
// Value cannot be used from concurrent goroutines.
// Use per-goroutine parsers or ParserPool instead.
type Value struct {
	o Object
	a []*Value
	s string
	t Type
}`,
}

var Conststr_3 string = `
type Type int

const (
	// TypeNull is JSON null.
	TypeNull Type = 0

	// TypeObject is JSON object type.
	TypeObject Type = 1

	// TypeArray is JSON array type.
	TypeArray Type = 2

	// TypeString is JSON string type.
	TypeString Type = 3

	// TypeNumber is JSON number type.
	TypeNumber Type = 4

	// TypeTrue is JSON true.
	TypeTrue Type = 5

	// TypeFalse is JSON false.
	TypeFalse Type = 6

	typeRawString Type = 7
)`
