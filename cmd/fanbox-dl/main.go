package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
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
	if v := os.Getenv("FANBOX_COOKIE"); v != "" {
		return v
	}

	return ""
}

var creatorFlag = &cli.StringFlag{
	Name:     "creator",
	Usage:    "Comma separated creator IDs to download. DO NOT prepend '@' to the creator ID.",
	Required: false,
}
var ignoreCreatorFlag = &cli.StringFlag{
	Name:     "ignore-creator",
	Usage:    "Comma separated creator IDs to ignore to download.",
	Required: false,
}
var sessIDFlag = &cli.StringFlag{
	Name:     "sessid",
	Usage:    "FANBOXSESSID which is stored in Cookies. If this is not set, fanbox-dl refers FANBOXSESSID environment value.",
	Required: false,
}
var cookieFlag = &cli.StringFlag{
	Name:     "cookie",
	Usage:    "Cookie for Fanbox API. This value overrides FANBOXSESSID environment value.",
	Required: false,
}
var userAgentFlag = &cli.StringFlag{
	Name:  "user-agent",
	Usage: "User-Agent for Fanbox API.",
	Value: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
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
	Usage: "Whether to check all posts. If --all=false, finish to crawling posts when found an already downloaded image.",
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
var skipOnErrorFlag = &cli.BoolFlag{
	Name:  "skip-on-error",
	Value: false,
	Usage: "Whether to skip downloading instead of exiting when an error occurred.",
}

var app = &cli.App{
	Name:  "fanbox-dl",
	Usage: "This CLI downloads images of supporting and following creators.",
	Flags: []cli.Flag{
		creatorFlag,
		ignoreCreatorFlag,
		sessIDFlag,
		cookieFlag,
		saveDirFlag,
		dirByPostFlag,
		dirByPlanFlag,
		userAgentFlag,
		allFlag,
		supportingFlag,
		followingFlag,
		skipFiles,
		dryRunFlag,
		verboseFlag,
		skipOnErrorFlag,
	},
	Action: func(c *cli.Context) error {
		initLogger(c.Bool(verboseFlag.Name))
		slog.Info("Launching Pixiv FANBOX Downloader!")

		var cookieStr string
		if sessID := resolveSessionID(c); sessID != "" {
			cookieStr = fmt.Sprintf("FANBOXSESSID=%s", sessID)
		}
		if v := c.String(cookieFlag.Name); v != "" {
			cookieStr = v
		}

		httpClient := retryablehttp.NewClient()
		httpClient.Logger = slog.Default()

		api := &fanbox.OfficialAPIClient{
			HTTPClient: httpClient,
			Cookie:     cookieStr,
			UserAgent:  c.String(userAgentFlag.Name),
		}

		client := &fanbox.Client{
			CheckAllPosts:     c.Bool(allFlag.Name),
			DryRun:            c.Bool(dryRunFlag.Name),
			SkipFiles:         c.Bool(skipFiles.Name),
			SkipOnError:       c.Bool(skipOnErrorFlag.Name),
			OfficialAPIClient: api,
			Storage: &fanbox.LocalStorage{
				SaveDir:   c.String(saveDirFlag.Name),
				DirByPost: c.Bool(dirByPostFlag.Name),
				DirByPlan: c.Bool(dirByPlanFlag.Name),
			},
		}

		ctx := c.Context
		startedAt := time.Now()

		idLister := &fanbox.CreatorIDLister{
			OfficialAPIClient: api,
		}

		in := &fanbox.CreatorIDListerDoInput{
			IncludeSupporting: c.Bool(supportingFlag.Name),
			IncludeFollowing:  c.Bool(followingFlag.Name),
		}
		if c.String(creatorFlag.Name) != "" {
			in.InputCreatorIDs = strings.Split(c.String(creatorFlag.Name), ",")
		}
		if c.String(ignoreCreatorFlag.Name) != "" {
			in.IgnoreCreatorIDs = strings.Split(c.String(ignoreCreatorFlag.Name), ",")
		}

		ids, err := idLister.Do(ctx, in)
		if err != nil {
			return fmt.Errorf("resolve creator IDs: %w", err)
		}
		for _, id := range ids {
			slog.InfoContext(ctx, "Start downloading", "creator_id", id)
			if err := client.Run(ctx, id); err != nil {
				return fmt.Errorf("failed downloading of %q: %w", id, err)
			}
		}

		slog.InfoContext(ctx, "Completed.", "duration", time.Since(startedAt).Round(time.Millisecond*100))
		return nil
	},
}

func main() {
	if err := run(); err != nil {
		log.Printf("%s ERROR LOG %s", strings.Repeat("=", 5), strings.Repeat("=", 5))
		log.Printf("fanbox-dl error: %s", err)
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

func initLogger(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	logger := slog.New(h)
	slog.SetDefault(logger)
}
