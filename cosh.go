package main

import (
	"github.com/brotherdetjr/goja"
)

func main() {
	vm := goja.New()
	global := vm.GlobalObject()
	global.Set("nunjucks", evalResource("embedded/nunjucks.js", vm))
	coffee := evalResource("embedded/coffee-script.js", vm)
	global.Set("coffee", coffee)
	compile := coffee.ToObject(vm).Get("compile").Export().(func(goja.FunctionCall) goja.Value)
	arg := goja.FunctionCall{
		This: vm.GlobalObject(),
		Arguments: []goja.Value{
			goja.NewStringValue("({x, y}).x = {x: 42}"),
		},
	}
	println(compile(arg).String())

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

/*func evalCoffee(script string) goja.Value {

}*/
