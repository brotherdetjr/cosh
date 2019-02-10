package main

import (
	"github.com/brotherdetjr/goja"
)

func main() {
	vm := goja.New()
	vm.GlobalObject().Set("nunjucks", evalResource("embedded/nunjucks.js", vm))
	r, _ := vm.RunString(`
	'Hello {{ username }}!'.render({ username: 'Disa' });
	`)
	println(r.String())

}

func evalResource(resourceName string, vm *goja.Runtime) goja.Value {
	if byteArray, err := Asset(resourceName); err != nil {
		panic(err)
	} else {
		if value, err := vm.RunString(string(byteArray[:])); err != nil {
			panic(err)
		} else {
			return value
		}
	}
}
