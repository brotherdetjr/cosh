package main

import (
	"bufio"
	"github.com/brotherdetjr/goja"
	"os/exec"
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

	cmd1 := exec.Command("echo", "Hello World!")
	cmd2 := exec.Command("sed", `s/o/x/g`)

	outPipe, err := cmd1.StdoutPipe()
	if err != nil {
		panic(err)
	}
	inPipe, err := cmd2.StdinPipe()
	if err != nil {
		panic(err)
	}
	resultPipe, err := cmd2.StdoutPipe()
	if err != nil {
		panic(err)
	}
	scanner1 := bufio.NewScanner(outPipe)
	scanner2 := bufio.NewScanner(resultPipe)

	go func() {
		for scanner1.Scan() {
			if _, err := inPipe.Write(scanner1.Bytes()); err != nil {
				panic(err)
			}
		}
		inPipe.Close()
	}()

	go func() {
		for scanner2.Scan() {
			print(scanner2.Text())
		}
	}()

	cmd2.Start()
	cmd1.Start()

	if err := cmd2.Wait(); err != nil {
		panic(err)
	}
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
