package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/yolo-pkgs/gore/binner"
)

func main() {
	app := &cli.App{
		Usage:                  `"npm list/update -g" for Go`,
		Suggest:                true,
		UseShortOptionHandling: true,
		Authors: []*cli.Author{
			{
				Name:  "Gleb Buzin",
				Email: "qufiwefefwoyn@gmail.com",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "list installed binaries",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "dev", Aliases: []string{"d"}, Usage: "also check dev packages"},
					&cli.BoolFlag{Name: "group", Aliases: []string{"g"}, Usage: "group packages by domain"},
					&cli.BoolFlag{Name: "extra", Aliases: []string{"e"}, Usage: "output extra info"},
					&cli.BoolFlag{Name: "simple", Aliases: []string{"s"}, Usage: "print without table"},
				},
				Action: func(c *cli.Context) error {
					binService, err := binner.New(c.Bool("simple"), c.Bool("dev"), c.Bool("extra"), c.Bool("group"))
					if err != nil {
						return err
					}

					return binService.ListBins()
				},
			},
			{
				Name:  "ls",
				Usage: "ls installed binaries",
				Action: func(_ *cli.Context) error {
					binService, err := binner.New(false, false, false, false)
					if err != nil {
						return err
					}

					return binService.LSBins()
				},
			},
			{
				Name:  "update",
				Usage: "update binaries",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "dev",
						Aliases: []string{"d"},
						Usage:   "check actual repos of dev packages and install them precisely",
					},
				},
				Action: func(c *cli.Context) error {
					binService, err := binner.New(false, c.Bool("dev"), false, false)
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
					binService, err := binner.New(false, false, false, false)
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
