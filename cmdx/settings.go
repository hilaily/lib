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

// ISet settings of cmdx
type ISet interface {
	IPrint

	CheckErr(err error, msg ...string)
	Throw(format string, a ...any)
	Recover(r any)
}

// IPrint ...
type IPrint interface {
	Normal(format string, a ...any)
	Green(format string, a ...any)
	Yellow(format string, a ...any)
	Red(format string, a ...any)
}

// DefaultSet ...
type DefaultSet struct {
	IPrint
}

// CheckErr ...
func (d *DefaultSet) CheckErr(err error, msg ...string) {
	if err != nil {
		if len(msg) > 0 {
			panic(&Err{error: fmt.Errorf("%s %w", msg[0], err)})
		}
		panic(err)
	}
}

// Throw ...
func (d *DefaultSet) Throw(format string, a ...any) {
	panic(&Err{error: fmt.Errorf(format, a...)})
}

// Recover ...
func (d *DefaultSet) Recover(r interface{}) {
	if r == nil {
		return
	}
	e, ok := r.(*Err)
	if ok {
		d.Red("%v", e.Error())
		return
	}
	panic(r)
}

// DefaultPrint ...
type DefaultPrint struct{}

// Normal ...
func (d *DefaultPrint) Normal(format string, a ...any) {
	fmt.Printf(format, a...)
}

// Green ...
func (d *DefaultPrint) Green(format string, a ...any) {
	color.Green(format, a...)
}

// Yellow ...
func (d *DefaultPrint) Yellow(format string, a ...any) {
	color.Yellow(format, a...)
}

// Red ...
func (d *DefaultPrint) Red(format string, a ...any) {
	color.Red(format, a...)
}

// Err ...
type Err struct {
	error
}
