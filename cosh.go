package main

import (
	"bufio"
	"fmt"
	"github.com/brotherdetjr/goja"
	"io"
	"os"
	"os/exec"
	"reflect"
	"sync"
)

var crlf = []byte{'\n'}

func main() {
	vm := goja.New()
	vm.SetTemplateRenderer(evalResource("embedded/nunjucks.js", vm), "renderString")
	vm.SetResolver(evalResource("embedded/resolver.js", vm).(*goja.Object))
	vm.Set("newScanLink", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(&ScanLink{
			Prev:      nil,
			Type:      Text, // TODO
			WaitGroup: sync.WaitGroup{},
			SplitFunc: bufio.ScanLines, // TODO
			Scanner:   nil,
			Writer:    nil,
			Mapping:   call.Arguments[0].Export().(func(call goja.FunctionCall) goja.Value),
			Runtime:   vm,
		})
	})
	// TODO
	mappingWrapper, _ := vm.RunString(`
(function() {
	var self;
	self = function (other) {
		return join(self, other);
	};
	self.link = newScanLink(arguments[0]);
	return self;
})
`)
	vm.Set("mappingWrapper", mappingWrapper)
	vm.Set("join", func(a *goja.Object, b *goja.Object) *goja.Object {
		if b.Get("link") == nil {
			arg := goja.FunctionCall{
				This:      vm.GlobalObject(),
				Arguments: []goja.Value{b},
			}
			b = vm.Get("mappingWrapper").Export().(func(goja.FunctionCall) goja.Value)(arg).(*goja.Object)
		}
		Join(extractLink(a), extractLink(b))
		return b
	})
	vm.Set("$", func(wrapped goja.Value) error {
		switch w := wrapped.(type) {
		case *goja.Object:
			return Consume(extractLink(w), os.Stdout)
		default:
			fmt.Println(w.String())
			return nil
		}
	})
	vm.Set("newCmdLink", func(call goja.FunctionCall) goja.Value {
		argCount := len(call.Arguments) - 1
		tail := make([]string, argCount)
		for key, value := range call.Arguments[:argCount] {
			tail[key] = value.String()
		}
		return vm.ToValue(&CmdLink{Cmd: exec.Command(call.Arguments[argCount].String(), tail...)})
	})
	coffee := evalResource("embedded/coffee-script.js", vm).
		ToObject(vm).
		Get("compile").
		Export().(func(goja.FunctionCall) goja.Value)
	cosh := `
$ (sh '-c', './mycmd.sh')((x) -> x + '?')(sed '-u', 's/o/0/g')
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

func Join(a Link, b Link) {
	b.SetPrevious(a)
	b.SetStdin(a.GetStdoutPipe())
}

func Consume(link Link, writer io.WriteCloser) error {
	link.SetStdout(writer)
	link.Start()
	last := link
	for link.GetPrevious() != nil {
		link = link.GetPrevious()
		link.Start()
	}
	return last.Wait()
}

type Link interface {
	SetStdin(reader io.Reader)
	SetStdout(writer io.WriteCloser)
	GetStdoutPipe() io.Reader
	SetPrevious(previous Link)
	GetPrevious() Link
	Start()
	Wait() error
}

type ScanType bool

const (
	Text   ScanType = false
	Binary ScanType = true
)

type ScanLink struct {
	Prev      Link
	Type      ScanType
	WaitGroup sync.WaitGroup
	SplitFunc bufio.SplitFunc
	Scanner   *bufio.Scanner
	Writer    io.WriteCloser
	Mapping   func(call goja.FunctionCall) goja.Value
	Runtime   *goja.Runtime
}

func (link *ScanLink) SetStdin(reader io.Reader) {
	link.Scanner = bufio.NewScanner(reader)
}

func (link *ScanLink) SetStdout(writer io.WriteCloser) {
	link.Writer = writer
}

func (link *ScanLink) GetStdoutPipe() io.Reader {
	pr, pw := io.Pipe()
	link.Writer = pw
	// TODO add to close list
	return pr
}

func (link *ScanLink) SetPrevious(previous Link) {
	link.Prev = previous
}

func (link *ScanLink) GetPrevious() Link {
	return link.Prev
}

func (link *ScanLink) Start() {
	link.WaitGroup.Add(1)
	go func() {
		scanner := link.Scanner
		for scanner.Scan() {
			var in interface{}
			if link.Type == Text {
				in = scanner.Text()
			} else {
				in = scanner.Bytes()
			}
			out := link.Mapping(goja.FunctionCall{
				This:      link.Runtime.ToValue(nil),
				Arguments: []goja.Value{link.Runtime.ToValue(in)},
			}).Export()
			buf := []byte(out.(string))
			if link.Type == Text {
			} else {
				// TODO
			}
			// TODO
			_, _ = link.Writer.Write(buf)
			if link.Type == Text &&
				reflect.ValueOf(link.SplitFunc).Pointer() == reflect.ValueOf(bufio.ScanLines).Pointer() {
				// TODO
				_, _ = link.Writer.Write(crlf)
			}
		}
		// TODO
		_ = link.Writer.Close()
		link.WaitGroup.Done()
	}()
}

func (link *ScanLink) Wait() error {
	link.WaitGroup.Wait()
	return nil
}

type CmdLink struct {
	Prev Link
	Cmd  *exec.Cmd
}

func (link *CmdLink) GetStdoutPipe() io.Reader {
	// TODO
	pipe, _ := link.Cmd.StdoutPipe()
	return pipe
}

func (link *CmdLink) SetStdin(reader io.Reader) {
	link.Cmd.Stdin = reader
}

func (link *CmdLink) SetStdout(writer io.WriteCloser) {
	link.Cmd.Stdout = writer
}

func (link *CmdLink) GetPrevious() Link {
	return link.Prev
}

func (link *CmdLink) SetPrevious(previous Link) {
	link.Prev = previous
}

func (link *CmdLink) Start() {
	// TODO
	_ = link.Cmd.Start()
}

func (link *CmdLink) Wait() error {
	return link.Cmd.Wait()
}
