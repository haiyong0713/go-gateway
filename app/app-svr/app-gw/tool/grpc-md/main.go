//nolint:biliautomaxprocs
package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {
	flag := &Flag{}
	app := cli.NewApp()
	app.Name = ""
	app.Usage = "grpc method信息生成工具"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "bapis",
			Usage:       "Specify the bapis project directory",
			Destination: &flag.Path,
		},
		cli.StringFlag{
			Name:        "p",
			Usage:       "Specifies the package name of the generated file",
			Destination: &flag.Package,
		},
		cli.StringFlag{
			Name:        "o",
			Usage:       "Specify the directory in which to generate the files. Output to standard output if not specified",
			Destination: &flag.OutPath,
		},
	}
	app.Action = func(ctx *cli.Context) error {
		return run(ctx, flag)
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
