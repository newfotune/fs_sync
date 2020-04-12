package main

import (
	"fmt"
	"fs_sync/watcher"
	"log"
	"os"
	"sort"

	"github.com/urfave/cli/v2"
)

func main() {
	w := watcher.New()

	defer w.Close()

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "setup",
				Aliases: []string{"s"},
				Value:   ".",
				Usage:   "Setup current directory",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "starts syncing to remote host",
				Action: func(c *cli.Context) error {
					w.Start()
					return nil
				},
			},
			{
				Name:  "add",
				Usage: "add a directory to be synced to remote host",
				Action: func(c *cli.Context) error {
					directories := c.Args().Slice()
					if len(directories) == 0 {
						return fmt.Errorf("no directories to sync")
					}

					return w.AddDirs(directories...)
				},
			},
			{
				Name:  "whitelist",
				Usage: "add a directory or file to not be synced",
				Action: func(c *cli.Context) error {
					log.Println("in no sync")
					return nil
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	done := make(chan bool)
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	<-done

}
