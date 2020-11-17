package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hareku/fanbox-dl/internal/download"
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
			Usage: "Directory for save images. (default './images/[user]')",
		},
		&cli.BoolFlag{
			Name:  "dir-by-post",
			Value: false,
			Usage: "Whether to separate save directories for each post",
		},
	}

	app.Action = func(c *cli.Context) error {
		fmt.Println("Launching Pixiv FANBOX Downloader!")
		fmt.Printf("Input User ID: %s\n", c.String("user"))

		saveDir := c.String("save-dir")
		if saveDir == "" {
			saveDir = filepath.Join("./images", c.String("user"))
		}

		client := download.Client{
			UserID:         c.String("user"),
			SaveDir:        saveDir,
			FANBOXSESSID:   c.String("sessid"),
			SeparateByPost: c.Bool("dir-by-post"),
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
		log.Fatal(err)
	}
	os.Exit(0)
}
