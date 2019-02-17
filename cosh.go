package main

import (
	"github.com/brotherdetjr/goja"
	"os"
	"os/exec"
)

func main() {
	vm := goja.New()
	vm.SetTemplateRenderer(evalResource("embedded/nunjucks.js", vm), "renderString")
	vm.SetResolver(evalResource("embedded/resolver.js", vm).(*goja.Object))
	coffee := evalResource("embedded/coffee-script.js", vm)
	compile := coffee.ToObject(vm).Get("compile").Export().(func(goja.FunctionCall) goja.Value)
	arg := goja.FunctionCall{
		This: vm.GlobalObject(),
		Arguments: []goja.Value{
			goja.NewStringValue(`
return echo 0, 1, 2, "{{ x }}{{ x }}".render(x: 33)
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
