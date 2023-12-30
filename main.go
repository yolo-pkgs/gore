package main

import (
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
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "simple", Aliases: []string{"s"}},
				},
				Action: func(c *cli.Context) error {
					binService, err := binner.New(c.Bool("simple"))
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
							binService, err := binner.New(false)
							if err != nil {
								return err
							}
							return binService.Update()
						},
					},
					{
						Name:  "dump",
						Usage: "dumps commands to install bins",
						Action: func(_ *cli.Context) error {
							binService, err := binner.New(false)
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
