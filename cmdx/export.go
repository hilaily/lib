package cmdx

import (
	"os"

	"github.com/urfave/cli/v2"
)

func WrapApp(app *cli.App) {
	defer func() {
		std.Recover(recover())
	}()

	err := app.Run(os.Args)
	std.CheckErr(err)
}

func Set(s ISet) {
	std = s
}

func CheckErr(err error, msg ...string) {
	std.CheckErr(err, msg...)
}

func Throw(format string, a ...any) {
	std.Throw(format, a...)
}

func Recover(r any) {
	std.Recover(r)
}

func Red(format string, a ...any) {
	std.Red(format, a...)
}

func Yellow(format string, a ...any) {
	std.Yellow(format, a...)
}
func Green(format string, a ...any) {
	std.Green(format, a...)
}

func Normal(format string, a ...any) {
	std.Normal(format, a...)
}
