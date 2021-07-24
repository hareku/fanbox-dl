package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/hareku/fanbox-dl/pkg/fanbox"
	"github.com/urfave/cli/v2"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetPrefix("[fanbox-dl] ")
}

func main() {
	app := &cli.App{
		Name:  "fanbox-dl",
		Usage: "Downloads all original images of a user.",
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     "user",
			Usage:    "Pixiv user ID to download, don't prepend '@'.",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "sessid",
			Usage:    "FANBOXSESSID which is stored in Cookies.",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "save-dir",
			Value: "./images",
			Usage: "The save destination folder",
		},
		&cli.BoolFlag{
			Name:  "dir-by-post",
			Value: false,
			Usage: "Whether to separate save directories for each post.",
		},
		&cli.BoolFlag{
			Name:  "all",
			Value: false,
			Usage: "Whether to check all posts. If --all=false, finish to download when found an already downloaded image.",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Value: false,
			Usage: "Whether to dry-run. in dry-run, not download images and output logs only.",
		},
	}

	app.Action = func(c *cli.Context) error {
		log.Print("Launching Pixiv FANBOX Downloader!")
		log.Printf("Input User ID: %q", c.String("user"))

		httpClient := fanbox.NewHTTPClientWithSession(c.String("sessid"))
		httpClient.Timeout = time.Second * 30

		storage := fanbox.NewLocalStorage(&fanbox.NewLocalStorageInput{
			SaveDir:   c.String("save-dir"),
			DirByPost: c.Bool("dir-by-post"),
		})

		client := fanbox.NewClient(&fanbox.NewClientInput{
			UserID:        c.String("user"),
			CheckAllPosts: c.Bool("all"),
			DryRun:        c.Bool("dry-run"),
			API:           fanbox.NewAPI(httpClient, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5)),
			Storage:       storage,
		})

		start := time.Now()
		err := client.Run(c.Context)
		if err != nil {
			return fmt.Errorf("download error: %w", err)
		}

		log.Printf("Completed (after %v).", time.Since(start))
		return nil
	}

	ctx := context.Background()
	err := app.RunContext(ctx, os.Args)
	if err != nil {
		log.Fatalf("Pixiv FANBOX Downloader failed: %s", err)
	}
	os.Exit(0)
}
