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

var app = &cli.App{
	Name:  "fanbox-dl",
	Usage: "This CLI downloads images of supporting and following creators.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "creator",
			Usage:    "Pixiv creator ID to download if you want to specify a creator. DO NOT prepend '@'.",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "sessid",
			Usage:    "FANBOXSESSID which is stored in Cookies.",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "save-dir",
			Value: "./images",
			Usage: "Directory to save images.",
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
			Usage: "Whether to dry-run. In dry-run, not download images and output logs only.",
		},
	},
	Action: func(c *cli.Context) error {
		log.Print("Launching Pixiv FANBOX Downloader!")
		startedAt := time.Now()

		httpClient := fanbox.NewHTTPClientWithSession(c.String("sessid"))
		httpClient.Timeout = time.Second * 30
		storage := fanbox.NewLocalStorage(&fanbox.NewLocalStorageInput{
			SaveDir:   c.String("save-dir"),
			DirByPost: c.Bool("dir-by-post"),
		})
		api := fanbox.NewAPI(httpClient, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5))
		client := fanbox.NewClient(&fanbox.NewClientInput{
			CheckAllPosts: c.Bool("all"),
			DryRun:        c.Bool("dry-run"),
			API:           api,
			Storage:       storage,
		})

		userID := c.String("user")
		if userID != "" {
			log.Printf("Input User ID: %q", userID)
			if err := client.Run(c.Context, userID); err != nil {
				return fmt.Errorf("download error: %w", err)
			}
		} else {
			plans, err := api.ListPlans(c.Context)
			if err != nil {
				return fmt.Errorf("failed to list plans: %w", err)
			}
			log.Printf("Found your %d supporting plans.", len(plans.Body))
			for _, p := range plans.Body {
				log.Printf("Start downloading of %q's images", p.CreatorID)
				if err := client.Run(c.Context, p.CreatorID); err != nil {
					return fmt.Errorf("download error: %w", err)
				}
			}
		}

		log.Printf("Completed (after %v).", time.Since(startedAt))
		return nil
	},
}

func main() {
	if err := app.RunContext(context.Background(), os.Args); err != nil {
		log.Fatalf("fanbox-dl failed to run: %s", err)
	}
	os.Exit(0)
}
