package main

import (
	"fmt"
	"github.com/brotherdetjr/goja"
	"io"
	"os"
	"os/exec"
)

func main() {
	vm := goja.New()
	vm.SetTemplateRenderer(evalResource("embedded/nunjucks.js", vm), "renderString")
	vm.SetResolver(evalResource("embedded/resolver.js", vm).(*goja.Object))
	vm.Set("join", func(a *goja.Object, b *goja.Object) {
		Join(extractLink(a), extractLink(b))
	})
	vm.Set("$", func(wrapped goja.Value) {
		switch w := wrapped.(type) {
		case *goja.Object:
			Consume(extractLink(w), os.Stdout)
		default:
			fmt.Println(w.String())
		}
	})
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
	cosh := `
$ (echo 'hello', 'world!!!')(sed 's/o/0/g')(sed 's/l/1/g')
$ 'Hello, {{ name }}!!!'.render name: 'Maximka'
`
	arg := goja.FunctionCall{
		This:      vm.GlobalObject(),
		Arguments: []goja.Value{vm.ToValue(cosh)},
	}
	js := coffee(arg).String()
	if _, err := vm.RunString(js); err != nil {
		panic(err)
	}
}

func extractLink(wrapped *goja.Object) Link {
	return wrapped.Get("link").Export().(Link)
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

func getLast(link Link) Link {
	for link.GetNext() != nil {
		link = link.GetNext()
	}
	return link
}

func Join(a Link, b Link) {
	last := getLast(a)
	last.SetNext(b)
	b.SetPrevious(last)
	b.SetStdin(last.GetStdoutPipe())
}

func Consume(link Link, writer io.Writer) {
	last := getLast(link)
	last.SetStdout(writer)
	last.Start()
	link = last
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
