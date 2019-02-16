package main

import (
	"github.com/brotherdetjr/goja"
	"os"
	"os/exec"
)

func main() {
	vm := goja.New()
	global := vm.GlobalObject()
	_ = global.Set("nunjucks", evalResource("embedded/nunjucks.js", vm))
	coffee := evalResource("embedded/coffee-script.js", vm)
	_ = global.Set("coffee", coffee)
	compile := coffee.ToObject(vm).Get("compile").Export().(func(goja.FunctionCall) goja.Value)
	arg := goja.FunctionCall{
		This: vm.GlobalObject(),
		Arguments: []goja.Value{
			goja.NewStringValue(`
f = (x, y) => x + y
[a, b] = [3, 4]
return "f({{ a }}, {{ b }}) = {{ f }}".render a: a, b: b, f: f(a, b) 
`),
		},
	}
	js := compile(arg).String()
	println(js)
	if value, err := vm.RunString(js); err != nil {
		panic(err)
	} else {
		println(value.String())
	}

	c1 := exec.Command("echo", "hello world\ndiso")
	c2 := exec.Command("sed", "s/o/x/g")

	c2.Stdin, _ = c1.StdoutPipe()
	c2.Stdout = os.Stdout
	_ = c2.Start()
	_ = c1.Start()
	_ = c2.Wait()
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
