package cmdx

import (
	"os"
)

type ICli interface {
	Run(arguments []string) (err error) // for urfave/cli
}

type ICobra interface {
	Execute() error
}

func WrapCli(cli ICli) {
	defer func() {
		std.Recover(recover())
	}()

	err := cli.Run(os.Args)
	std.CheckErr(err)
}

func WrapCobra(cobra ICobra) {
	defer func() {
		std.Recover(recover())
	}()

	err := cobra.Execute()
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
