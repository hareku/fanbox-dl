package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/hareku/fanbox-dl/pkg/fanbox"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/urfave/cli/v2"
)

func resolveSessionID(c *cli.Context) string {
	if v := c.String(sessIDFlag.Name); v != "" {
		return v
	}

	if v := os.Getenv("FANBOXSESSID"); v != "" {
		return v
	}

	return ""
}

var creatorFlag = &cli.StringFlag{
	Name:     "creator",
	Usage:    "Pixiv creator ID to download if you want to specify a creator. DO NOT prepend '@'.",
	Required: false,
}
var sessIDFlag = &cli.StringFlag{
	Name:     "sessid",
	Usage:    "FANBOXSESSID which is stored in Cookies. If this is not set, fanbox-dl refers FANBOXSESSID environment value.",
	Required: false,
}
var saveDirFlag = &cli.StringFlag{
	Name:  "save-dir",
	Value: "./images",
	Usage: "Directory to save images.",
}
var dirByPostFlag = &cli.BoolFlag{
	Name:  "dir-by-post",
	Value: false,
	Usage: "Whether to separate save directories by post title.",
}
var dirByPlanFlag = &cli.BoolFlag{
	Name:  "dir-by-plan",
	Value: false,
	Usage: "Whether to separate save directories by plan.",
}
var allFlag = &cli.BoolFlag{
	Name:  "all",
	Value: false,
	Usage: "Whether to check all posts. If --all=false, finish to download when found an already downloaded image.",
}
var supportingFlag = &cli.BoolFlag{
	Name:  "supporting",
	Value: true,
	Usage: "Whether to download images of supporting creators.",
}
var followingFlag = &cli.BoolFlag{
	Name:  "following",
	Value: true,
	Usage: "Whether to download images of following creators.",
}
var skipFiles = &cli.BoolFlag{
	Name:  "skip-files",
	Value: false,
	Usage: "Whether to skip downloading files (not images).",
}
var dryRunFlag = &cli.BoolFlag{
	Name:  "dry-run",
	Value: false,
	Usage: "Whether to dry-run. In dry-run, fanbox-dl skip downloading files.",
}
var verboseFlag = &cli.BoolFlag{
	Name:  "verbose",
	Value: false,
	Usage: "Whether to output debug logs.",
}

var app = &cli.App{
	Name:  "fanbox-dl",
	Usage: "This CLI downloads images of supporting and following creators.",
	Flags: []cli.Flag{
		creatorFlag,
		sessIDFlag,
		saveDirFlag,
		dirByPostFlag,
		dirByPlanFlag,
		allFlag,
		supportingFlag,
		followingFlag,
		skipFiles,
		dryRunFlag,
		verboseFlag,
	},
	Action: func(c *cli.Context) error {
		logger := fanbox.NewLogger(&fanbox.NewLoggerInput{
			Out:     os.Stdout,
			Verbose: c.Bool(verboseFlag.Name),
		})
		logger.Infof("Launching Pixiv FANBOX Downloader!")

		sessID := resolveSessionID(c)
		if sessID == "" {
			logger.Infof("Fanbox SessionID is not set. Starting as a guest.")
		}

		httpClient := retryablehttp.NewClient()
		httpClient.Logger = logger

		api := &fanbox.OfficialAPIClient{
			HTTPClient: httpClient,
			SessionID:  sessID,
		}

		client := &fanbox.Client{
			CheckAllPosts:     c.Bool(allFlag.Name),
			DryRun:            c.Bool(dryRunFlag.Name),
			SkipFiles:         c.Bool(skipFiles.Name),
			OfficialAPIClient: api,
			Storage: &fanbox.LocalStorage{
				SaveDir:   c.String(saveDirFlag.Name),
				DirByPost: c.Bool(dirByPostFlag.Name),
				DirByPlan: c.Bool(dirByPlanFlag.Name),
			},
			Logger: logger,
		}

		ctx := c.Context
		startedAt := time.Now()

		idLister := &fanbox.CreatorIDLister{
			OfficialAPIClient: api,
		}
		ids, err := idLister.Do(ctx, &fanbox.CreatorIDListerDoInput{
			InputCreatorID:    c.String(creatorFlag.Name),
			IncludeSupporting: c.Bool(supportingFlag.Name),
			IncludeFollowing:  c.Bool(followingFlag.Name),
		})
		if err != nil {
			return fmt.Errorf("failed to resolve creator IDs: %w", err)
		}
		for _, id := range ids {
			logger.Infof("Started downloading of %q.", id)
			if err := client.Run(ctx, id); err != nil {
				return fmt.Errorf("failed downloading of %q: %w", id, err)
			}
		}

		logger.Infof("Completed (after %v).", time.Since(startedAt))
		return nil
	},
}

func main() {
	if err := run(); err != nil {
		log.Printf("%s ERROR LOG %s", strings.Repeat("=", 5), strings.Repeat("=", 5))
		log.Printf("fanbox-dl failed to run: %s", err)
		log.Println(strings.Repeat("=", 21))

		log.Printf("The error log seems a bug, please open an issue on GitHub: %s.", "https://github.com/hareku/fanbox-dl/issues")
	}
	os.Exit(0)
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := app.RunContext(ctx, os.Args); err != nil {
		return err
	}
	return nil
}
