package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "patch",
				Aliases: []string{"p"},
				Usage:   "publish new patch version",
				Action: func(cCtx *cli.Context) error {
					return patch()
				},
			},
			{
				Name:    "bins",
				Aliases: []string{"t"},
				Usage:   "list installed binaries",
				Action: func(cCtx *cli.Context) error {
					return listBins()
				},
				Subcommands: []*cli.Command{
					{
						Name:  "update",
						Usage: "add a new template",
						Action: func(cCtx *cli.Context) error {
							fmt.Println("bins update")
							return nil
						},
					},
					{
						Name:  "dump",
						Usage: "dumps commands to install bins",
						Action: func(cCtx *cli.Context) error {
							return binDump()
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Panic(err)
	}
}
