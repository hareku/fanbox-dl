package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "fanboxdl",
		Usage: "Downloads all posted original images of the specified user.",
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     "user",
			Usage:    "Pixiv user ID, don't prepend @",
			Required: true,
		},
	}

	app.Action = func(c *cli.Context) error {
		fmt.Println("Launching Pixiv Fanbox Downloader!")
		fmt.Printf("Input User ID: %s\n", c.String("user"))
		return nil
	}

	ctx := context.Background()
	err := app.RunContext(ctx, os.Args)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
