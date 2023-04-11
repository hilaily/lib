package cmdx

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	std ISet = &DefaultSet{
		IPrint: &DefaultPrint{},
	}
)

type ISet interface {
	IPrint

	CheckErr(err error)
	Throw(format string, a ...any)
	Recover(r any)
}

type IPrint interface {
	Normal(format string, a ...any)
	Green(format string, a ...any)
	Yellow(format string, a ...any)
	Red(format string, a ...any)
}

type DefaultSet struct {
	IPrint
}

func (d *DefaultSet) CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (d *DefaultSet) Throw(format string, a ...any) {
	panic(fmt.Errorf(format, a...))
}

func (d *DefaultSet) Recover(r interface{}) {
	if r != nil {
		d.Red("%v", r)
	}
}

type DefaultPrint struct{}

func (d *DefaultPrint) Normal(format string, a ...any) {
	fmt.Printf(format, a...)
}

func (d *DefaultPrint) Green(format string, a ...any) {
	color.Green(format, a...)
}

func (d *DefaultPrint) Yellow(format string, a ...any) {
	color.Yellow(format, a...)
}

func (d *DefaultPrint) Red(format string, a ...any) {
	color.Red(format, a...)
}
