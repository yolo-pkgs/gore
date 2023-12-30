package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/yolo-pkgs/gore/binner"
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
				Action: func(_ *cli.Context) error {
					binService, err := binner.New()
					if err != nil {
						return err
					}
					return binService.ListBins()
				},
				Subcommands: []*cli.Command{
					{
						Name:  "update",
						Usage: "add a new template",
						Action: func(_ *cli.Context) error {
							fmt.Println("bins update")
							return nil
						},
					},
					{
						Name:  "dump",
						Usage: "dumps commands to install bins",
						Action: func(_ *cli.Context) error {
							binService, err := binner.New()
							if err != nil {
								return err
							}
							return binService.Dump()
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
