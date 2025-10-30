package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/hareku/fanbox-dl/internal/applog"
	"github.com/hareku/fanbox-dl/internal/tlsclient"
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

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionFlag = &cli.BoolFlag{
	Name:  "version",
	Value: false,
	Usage: "Print the version and exit.",
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
	Usage:    "Cookie for Fanbox API. This value overrides FANBOXSESSID.",
	Required: false,
}
var userAgentFlag = &cli.StringFlag{
	Name:  "user-agent",
	Usage: "User-Agent for Fanbox API.",
	Value: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
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
var skipImages = &cli.BoolFlag{
	Name:  "skip-images",
	Value: false,
	Usage: "Whether to skip downloading images.",
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
var removeUnprintableCharsFlag = &cli.BoolFlag{
	Name:  "remove-unprintable-chars",
	Value: false,
	Usage: "Whether to remove unprintable characters from file names.",
}

var app = &cli.App{
	Name:  "fanbox-dl",
	Usage: "This CLI downloads images of supporting and following creators.",
	Flags: []cli.Flag{
		versionFlag,
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
		skipImages,
		dryRunFlag,
		verboseFlag,
		skipOnErrorFlag,
		removeUnprintableCharsFlag,
	},
	Action: func(c *cli.Context) error {
		applog.InitLogger(c.Bool(verboseFlag.Name))
		slog.Info("Launching Pixiv FANBOX Downloader!", "version", version, "commit", commit, "date", date)
		if c.Bool(versionFlag.Name) {
			return nil
		}

		var cookieStr string
		if sessID := resolveSessionID(c); sessID != "" {
			slog.Debug("Using session ID", "sessid_bytes", len(sessID))
			cookieStr = fmt.Sprintf("FANBOXSESSID=%s", sessID)
		}
		if v := c.String(cookieFlag.Name); v != "" {
			if cookieStr != "" {
				slog.Warn("session ID and cookie are set, cookie option overrides session ID option")
			}
			slog.Debug("Using cookie", "cookie_bytes", len(v))
			cookieStr = v
		}

		httpClient := retryablehttp.NewClient()
		httpClient.Logger = slog.Default()
		httpClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
			if err != nil {
				return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
			}
			b, err := fanbox.IsFailedToThumbnailingErr(resp)
			if err == nil && b {
				return false, fanbox.ErrFailedToThumbnailing
			}
			return retryablehttp.DefaultRetryPolicy(ctx, resp, nil)
		}

		tlsTransp, err := tlsclient.NewTransportWithOptions(tls_client.NewNoopLogger(), tls_client.WithClientProfile(profiles.Chrome_131))
		if err != nil {
			return fmt.Errorf("create tls transport: %w", err)
		}
		httpClient.HTTPClient.Transport = tlsTransp

		api := &fanbox.OfficialAPIClient{
			HTTPClient: httpClient,
			Cookie:     cookieStr,
			UserAgent:  c.String(userAgentFlag.Name),
		}

		client := &fanbox.Client{
			CheckAllPosts:     c.Bool(allFlag.Name),
			DryRun:            c.Bool(dryRunFlag.Name),
			SkipFiles:         c.Bool(skipFiles.Name),
			SkipImages:        c.Bool(skipImages.Name),
			SkipOnError:       c.Bool(skipOnErrorFlag.Name),
			OfficialAPIClient: api,
			Storage: &fanbox.LocalStorage{
				SaveDir:   c.String(saveDirFlag.Name),
				DirByPost: c.Bool(dirByPostFlag.Name),
				DirByPlan: c.Bool(dirByPlanFlag.Name),

				RemoveUnprintableChars: c.Bool(removeUnprintableCharsFlag.Name),
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
		slog.Error("fanbox-dl Error", "error", err)
		slog.Error("The error log seems a bug, please open an issue on GitHub", "url", "https://github.com/hareku/fanbox-dl/issues")

		if errors.Is(err, fanbox.ErrStatusForbidden) {
			slog.Error("This 403 error may occur when connecting from an IP address outside of Japan. Please try again from VPN or other IP addresses in Japan.")
		}
		os.Exit(1)
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
