package main

import (
	"github.com/brotherdetjr/goja"
	"io"
	"os"
	"os/exec"
)

func main() {
	vm := goja.New()
	vm.SetTemplateRenderer(evalResource("embedded/nunjucks.js", vm), "renderString")
	vm.SetResolver(evalResource("embedded/resolver.js", vm).(*goja.Object))
	vm.Set("newCmdLink", func(call goja.FunctionCall) goja.Value {
		argCount := len(call.Arguments) - 1
		tail := make([]string, argCount)
		for key, value := range call.Arguments[:argCount] {
			tail[key] = value.String()
		}
		return vm.ToValue(NewCmdLink(call.Arguments[argCount].String(), tail...))
	})
	coffee := evalResource("embedded/coffee-script.js", vm).
		ToObject(vm).
		Get("compile").
		Export().(func(goja.FunctionCall) goja.Value)
	arg := goja.FunctionCall{
		This: vm.GlobalObject(),
		Arguments: []goja.Value{
			vm.ToValue(`
return newCmdLink('hello', 'world', 'echo').Cmd.Args
`),
		},
	}
	js := coffee(arg).String()
	println(js)
	if value, err := vm.RunString(js); err != nil {
		panic(err)
	} else {
		println(value.String())
	}

	c1 := NewCmdLink("echo", "hello world\ndiso")
	c2 := NewCmdLink("sed", "s/o/0/g")
	c3 := NewCmdLink("sed", "s/l/1/g")
	Join(c1, c2)
	Join(c2, c3)
	Consume(c1, os.Stdout)
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

func Join(a Link, b Link) {
	a.SetNext(b)
	b.SetPrevious(a)
	b.SetStdin(a.GetStdoutPipe())
}

func Consume(link Link, writer io.Writer) {
	for link.GetNext() != nil {
		link = link.GetNext()
	}
	last := link
	last.SetStdout(writer)
	last.Start()
	for link.GetPrevious() != nil {
		link = link.GetPrevious()
		link.Start()
	}
	last.Wait()
}

func NewCmdLink(name string, arg ...string) *CmdLink {
	return &CmdLink{Cmd: exec.Command(name, arg...)}
}

type Link interface {
	SetStdin(reader io.Reader)
	SetStdout(writer io.Writer)
	GetStdoutPipe() io.Reader
	SetNext(link Link)
	GetNext() Link
	SetPrevious(link Link)
	GetPrevious() Link
	Start()
	Wait()
}

type Bidirectional struct {
	Next Link
	Prev Link
}

type CmdLink struct {
	Bidirectional
	Cmd *exec.Cmd
}

func (cmdLink *CmdLink) GetStdoutPipe() io.Reader {
	pipe, _ := cmdLink.Cmd.StdoutPipe()
	return pipe
}

func (cmdLink *CmdLink) SetNext(link Link) {
	cmdLink.Next = link
}

func (cmdLink *CmdLink) SetStdin(reader io.Reader) {
	cmdLink.Cmd.Stdin = reader
}

func (cmdLink *CmdLink) SetStdout(writer io.Writer) {
	cmdLink.Cmd.Stdout = writer
}

func (cmdLink *CmdLink) GetNext() Link {
	return cmdLink.Next
}

func (cmdLink *CmdLink) GetPrevious() Link {
	return cmdLink.Prev
}

func (cmdLink *CmdLink) SetPrevious(link Link) {
	cmdLink.Prev = link
}

func (cmdLink *CmdLink) Start() {
	_ = cmdLink.Cmd.Start()
}

func (cmdLink *CmdLink) Wait() {
	_ = cmdLink.Cmd.Wait()
}
