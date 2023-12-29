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
				Usage:   "increase patch version",
				Action: func(cCtx *cli.Context) error {
					return patch()
				},
			},
			{
				Name:    "bins",
				Aliases: []string{"t"},
				Usage:   "options for task templates",
				Action: func(cCtx *cli.Context) error {
					fmt.Println("bins")
					return nil
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
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Panic(err)
	}
}
