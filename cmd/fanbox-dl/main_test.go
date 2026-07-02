package main

import (
	"flag"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestResolveCookie(t *testing.T) {
	tests := []struct {
		name       string
		sessIDFlag string
		cookieFlag string
		envSessID  string
		envCookie  string
		want       string
	}{
		{
			name:       "sessid flag becomes cookie",
			sessIDFlag: "flag-session",
			want:       "FANBOXSESSID=flag-session",
		},
		{
			name:      "FANBOXSESSID becomes cookie",
			envSessID: "env-session",
			want:      "FANBOXSESSID=env-session",
		},
		{
			name:      "FANBOX_COOKIE is full cookie",
			envCookie: "FANBOXSESSID=env-session; other=value",
			want:      "FANBOXSESSID=env-session; other=value",
		},
		{
			name:       "FANBOX_COOKIE overrides sessid",
			sessIDFlag: "flag-session",
			envCookie:  "FANBOXSESSID=env-session",
			want:       "FANBOXSESSID=env-session",
		},
		{
			name:       "cookie flag wins",
			sessIDFlag: "flag-session",
			cookieFlag: "FANBOXSESSID=cookie-flag; other=value",
			envCookie:  "FANBOXSESSID=env-session",
			want:       "FANBOXSESSID=cookie-flag; other=value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("FANBOXSESSID", tt.envSessID)
			t.Setenv("FANBOX_COOKIE", tt.envCookie)

			c := newTestCLIContext(t, tt.sessIDFlag, tt.cookieFlag)
			if got := resolveCookie(c); got != tt.want {
				t.Fatalf("resolveCookie() = %q, want %q", got, tt.want)
			}
		})
	}
}

func newTestCLIContext(t *testing.T, sessID string, cookie string) *cli.Context {
	t.Helper()
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.String(sessIDFlag.Name, "", "")
	set.String(cookieFlag.Name, "", "")
	if sessID != "" {
		if err := set.Set(sessIDFlag.Name, sessID); err != nil {
			t.Fatalf("Set(sessid) error = %v", err)
		}
	}
	if cookie != "" {
		if err := set.Set(cookieFlag.Name, cookie); err != nil {
			t.Fatalf("Set(cookie) error = %v", err)
		}
	}
	return cli.NewContext(cli.NewApp(), set, nil)
}
