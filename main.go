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
				Name:    "list",
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
			},
			{
				Name:  "update",
				Usage: "update binaries",
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
	}

	if err := app.Run(os.Args); err != nil {
		log.Panic(err)
	}
}
