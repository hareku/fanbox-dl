package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hareku/fanbox-dl/pkg/download"
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
			Usage:    "Pixiv user ID to download, don't prepend @",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "sessid",
			Usage:    "FANBOXSESSID which is stored in Cookies",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "save-dir",
			Value: "./images",
			Usage: "Directory for save images.",
		},
		&cli.BoolFlag{
			Name:  "dir-by-post",
			Value: false,
			Usage: "Whether to separate save directories for each post",
		},
		&cli.BoolFlag{
			Name:  "all",
			Value: false,
			Usage: "Whether to check all posts.",
		},
	}

	app.Action = func(c *cli.Context) error {
		fmt.Println("Launching Pixiv FANBOX Downloader!")
		fmt.Printf("Input User ID: %s\n", c.String("user"))

		client := download.Client{
			UserID:         c.String("user"),
			SaveDir:        c.String("save-dir"),
			FANBOXSESSID:   c.String("sessid"),
			SeparateByPost: c.Bool("dir-by-post"),
			CheckAllPosts:  c.Bool("all"),
		}
		err := client.Run(c.Context)
		if err != nil {
			return fmt.Errorf("download error: %w", err)
		}
		return nil
	}

	ctx := context.Background()
	err := app.RunContext(ctx, os.Args)
	if err != nil {
		log.Fatalf("Pixiv FANBOX Downloader failed: %s\n", err)
	}
	log.Println("Pixiv FANBOX Downloader completed.")
	os.Exit(0)
}
