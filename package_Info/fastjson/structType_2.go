package fastjsonPackageInfo

var StructMap_2 map[string]string = map[string]string{
	"Parser": `
type Parser struct {
	b []byte
	c cache
}`,
	"cache": `type cache struct {
	vs []Value
}`,
	"kv": `type kv struct {
	k string
	v *Value
}`,
	"Object": `
type Object struct {
	kvs           []kv
	keysUnescaped bool
}`,
	"Value": `
type Value struct {
	o Object
	a []*Value
	s string
	t Type
}`,
}

var Conststr_2 string = `
type Type int

const (
	TypeNull Type = 0

	TypeObject Type = 1

	TypeArray Type = 2

	TypeString Type = 3

	TypeNumber Type = 4

	TypeTrue Type = 5

	TypeFalse Type = 6

	typeRawString Type = 7
)`
