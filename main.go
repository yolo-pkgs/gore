package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/yolo-pkgs/gore/binner"
)

func main() {
	app := &cli.App{
		Usage:    `"npm list/update -g" for Go`,
		HideHelp: true,
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "list installed binaries",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "simple", Aliases: []string{"s"}},
					&cli.BoolFlag{Name: "dev", Aliases: []string{"d"}},
				},
				Action: func(c *cli.Context) error {
					binService, err := binner.New(c.Bool("simple"), c.Bool("dev"))
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
					binService, err := binner.New(false, false)
					if err != nil {
						return err
					}

					return binService.Update()
				},
			},
			{
				Name:  "dump",
				Usage: "dumps commands to install bins",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "latest", Aliases: []string{"l"}},
				},
				Action: func(c *cli.Context) error {
					binService, err := binner.New(false, false)
					if err != nil {
						return err
					}

					return binService.Dump(c.Bool("latest"))
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Panic(err)
	}
}
