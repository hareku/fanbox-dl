package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/hareku/fanbox-dl/pkg/fanbox"
	"github.com/urfave/cli/v2"
)

func resolveSessionID(c *cli.Context) string {
	if v := c.String("sessid"); v != "" {
		return v
	}

	if v := os.Getenv("FANBOXSESSID"); v != "" {
		return v
	}

	return ""
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
			Usage:    "FANBOXSESSID which is stored in Cookies. If this is not set, fanbox-dl refers FANBOXSESSID environment value.",
			Required: false,
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
			Name:  "supporting",
			Value: true,
			Usage: "Whether to download images of supporting creators.",
		},
		&cli.BoolFlag{
			Name:  "following",
			Value: true,
			Usage: "Whether to download images of following creators.",
		},
		&cli.BoolFlag{
			Name:  "with-files",
			Value: false,
			Usage: "Whether to download files creator uploaded (not images).",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Value: false,
			Usage: "Whether to dry-run. In dry-run, not download images and output logs only.",
		},
		&cli.BoolFlag{
			Name:  "verbose",
			Value: false,
			Usage: "Whether to output debug logs.",
		},
	},
	Action: func(c *cli.Context) error {
		logger := fanbox.NewLogger(&fanbox.NewLoggerInput{
			Out:     os.Stdout,
			Verbose: c.Bool("verbose"),
		})
		logger.Infof("Launching Pixiv FANBOX Downloader!")

		ctx := c.Context
		startedAt := time.Now()

		sessID := resolveSessionID(c)
		if sessID == "" {
			return errors.New("please set FANBOXSESSID to 'sessid' option or environment value, see --help")
		}

		httpClient := fanbox.NewHTTPClientWithSession(sessID)
		httpClient.Timeout = time.Second * 30
		api := fanbox.NewAPI(httpClient, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5))
		client := fanbox.NewClient(&fanbox.NewClientInput{
			CheckAllPosts: c.Bool("all"),
			DryRun:        c.Bool("dry-run"),
			DownloadFiles: c.Bool("with-files"),
			API:           api,
			Storage: fanbox.NewLocalStorage(&fanbox.NewLocalStorageInput{
				SaveDir:   c.String("save-dir"),
				DirByPost: c.Bool("dir-by-post"),
			}),
			FileStorage: fanbox.NewLocalFileStorage(&fanbox.NewLocalFileStorageInput{
				SaveDir:   c.String("save-dir"),
				DirByPost: c.Bool("dir-by-post"),
			}),
			Logger: logger,
		})

		resolver := fanbox.NewCreatorResolver(api)
		ids, err := resolver.Do(ctx, &fanbox.CreatorResolverDoInput{
			InputCreatorID:    fanbox.CreatorID(c.String("creator")),
			IncludeSupporting: c.Bool("supporting"),
			IncludeFollowing:  c.Bool("following"),
		})
		if err != nil {
			return fmt.Errorf("failed to resolve creator IDs: %w", err)
		}
		for _, id := range ids {
			logger.Infof("Started downloading images of %q.", id)
			if err := client.Run(ctx, string(id)); err != nil {
				return fmt.Errorf("failed to download image of %q: %w", id, err)
			}
		}

		logger.Infof("Completed (after %v).", time.Since(startedAt))
		return nil
	},
}

func main() {
	if err := app.RunContext(context.Background(), os.Args); err != nil {
		log.Println(fmt.Sprintf("%s ERROR LOG %s", strings.Repeat("=", 5), strings.Repeat("=", 5)))
		log.Printf("fanbox-dl failed to run: %s", err)
		log.Println(strings.Repeat("=", 21))

		log.Printf("The error log seems to a bug, please open an issue on GitHub: %s.", "https://github.com/hareku/fanbox-dl/issues")
	}
	os.Exit(0)
}
