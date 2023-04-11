package cmdx

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	std ISet = &DefaultSet{}
)

type ISet interface {
	Normal(format string, a ...any)
	Green(format string, a ...any)
	Yellow(format string, a ...any)
	Red(format string, a ...any)

	CheckErr(err error)
	Throw(format string, a ...any)
	Recover(r interface{})
}

func Set(s ISet) {
	std = s
}

type DefaultSet struct{}

func (d *DefaultSet) Normal(format string, a ...any) {
	fmt.Printf(format, a...)
}

func (d *DefaultSet) Green(format string, a ...any) {
	color.Green(format, a...)
}

func (d *DefaultSet) Yellow(format string, a ...any) {
	color.Yellow(format, a...)
}

func (d *DefaultSet) Red(format string, a ...any) {
	color.Red(format, a...)
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
		color.Red("%v", r)
	}
}
