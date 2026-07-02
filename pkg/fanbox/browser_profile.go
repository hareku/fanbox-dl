package fanbox

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/hareku/fanbox-dl/internal/tlsclient"
)

const (
	BrowserProfileAuto    = "auto"
	BrowserProfileChrome  = "chrome"
	BrowserProfileFirefox = "firefox"
)

type BrowserFamily string

const (
	BrowserFamilyChrome  BrowserFamily = "chrome"
	BrowserFamilyFirefox BrowserFamily = "firefox"
)

type BrowserProfile struct {
	Family                BrowserFamily
	TLSProfile            profiles.ClientProfile
	TLSProfileName        string
	HeaderOrder           []string
	AssetHeaderOrder      []string
	PseudoHeaderOrder     []string
	requestedBrowserMajor int
}

var browserVersionPattern = regexp.MustCompile(`(?i)(?:Chrome|Firefox)/(\d+)`)

func ResolveBrowserProfile(userAgent, requestedProfile string) (*BrowserProfile, error) {
	switch strings.ToLower(strings.TrimSpace(requestedProfile)) {
	case "", BrowserProfileAuto:
		if isFirefoxUserAgent(userAgent) {
			return newFirefoxProfile(userAgent), nil
		}
		return newChromeProfile(userAgent), nil
	case BrowserProfileChrome:
		return newChromeProfile(userAgent), nil
	case BrowserProfileFirefox:
		return newFirefoxProfile(userAgent), nil
	default:
		return nil, fmt.Errorf("invalid browser profile %q: expected auto, chrome, or firefox", requestedProfile)
	}
}

func (p *BrowserProfile) ApplyAPIHeaders(h http.Header, cookie string, userAgent string) {
	switch p.Family {
	case BrowserFamilyFirefox:
		applyFirefoxAPIHeaders(h, cookie, userAgent)
	default:
		applyChromeAPIHeaders(h, cookie, userAgent)
	}
	tlsclient.SetHeaderOrder(h, p.HeaderOrder, p.PseudoHeaderOrder)
}

func (p *BrowserProfile) ApplyAssetHeaders(h http.Header, cookie string, userAgent string) {
	switch p.Family {
	case BrowserFamilyFirefox:
		applyFirefoxAssetHeaders(h, cookie, userAgent)
	default:
		applyChromeAssetHeaders(h, cookie, userAgent)
	}
	tlsclient.SetHeaderOrder(h, p.AssetHeaderOrder, p.PseudoHeaderOrder)
}

func (p *BrowserProfile) RequestedBrowserMajor() int {
	return p.requestedBrowserMajor
}

func newChromeProfile(userAgent string) *BrowserProfile {
	return &BrowserProfile{
		Family:                BrowserFamilyChrome,
		TLSProfile:            profiles.Chrome_133,
		TLSProfileName:        "chrome_133",
		HeaderOrder:           chromeAPIHeaderOrder,
		AssetHeaderOrder:      chromeAssetHeaderOrder,
		PseudoHeaderOrder:     []string{":method", ":authority", ":scheme", ":path"},
		requestedBrowserMajor: parseBrowserMajor(userAgent),
	}
}

func newFirefoxProfile(userAgent string) *BrowserProfile {
	return &BrowserProfile{
		Family:                BrowserFamilyFirefox,
		TLSProfile:            profiles.Firefox_135,
		TLSProfileName:        "firefox_135",
		HeaderOrder:           firefoxAPIHeaderOrder,
		AssetHeaderOrder:      firefoxAssetHeaderOrder,
		PseudoHeaderOrder:     []string{":method", ":authority", ":scheme", ":path"},
		requestedBrowserMajor: parseBrowserMajor(userAgent),
	}
}

func isFirefoxUserAgent(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	return strings.Contains(ua, "firefox/")
}

func parseBrowserMajor(userAgent string) int {
	matches := browserVersionPattern.FindStringSubmatch(userAgent)
	if len(matches) != 2 {
		return 0
	}
	var major int
	_, _ = fmt.Sscanf(matches[1], "%d", &major)
	return major
}

func applyChromeAPIHeaders(h http.Header, cookie string, userAgent string) {
	setCookieHeader(h, cookie)
	h.Set("Origin", "https://www.fanbox.cc")
	h.Set("Referer", "https://www.fanbox.cc/")
	h.Set("User-Agent", userAgent)
	h.Set("Accept", "application/json, text/plain, */*")
	h.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	h.Set("Accept-Language", "en-US,en;q=0.9")
	h.Set("Priority", "u=1, i")
	h.Set("Sec-Ch-Ua", chromeSecCHUA(userAgent))
	h.Set("Sec-Ch-Ua-Mobile", "?0")
	h.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	h.Set("Sec-Fetch-Dest", "empty")
	h.Set("Sec-Fetch-Mode", "cors")
	h.Set("Sec-Fetch-Site", "same-site")
}

func applyChromeAssetHeaders(h http.Header, cookie string, userAgent string) {
	setCookieHeader(h, cookie)
	h.Set("Referer", "https://www.fanbox.cc/")
	h.Set("User-Agent", userAgent)
	h.Set("Accept", "*/*")
	h.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	h.Set("Accept-Language", "en-US,en;q=0.9")
	h.Set("Priority", "u=1, i")
	h.Set("Sec-Ch-Ua", chromeSecCHUA(userAgent))
	h.Set("Sec-Ch-Ua-Mobile", "?0")
	h.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	h.Set("Sec-Fetch-Dest", "empty")
	h.Set("Sec-Fetch-Mode", "cors")
	h.Set("Sec-Fetch-Site", "same-site")
}

func applyFirefoxAPIHeaders(h http.Header, cookie string, userAgent string) {
	setCookieHeader(h, cookie)
	h.Set("Origin", "https://www.fanbox.cc")
	h.Set("Referer", "https://www.fanbox.cc/")
	h.Set("User-Agent", userAgent)
	h.Set("Accept", "application/json, text/plain, */*")
	h.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	h.Set("Accept-Language", "en-US,en;q=0.5")
	h.Set("Priority", "u=4")
	h.Set("Sec-Fetch-Dest", "empty")
	h.Set("Sec-Fetch-Mode", "cors")
	h.Set("Sec-Fetch-Site", "same-site")
}

func applyFirefoxAssetHeaders(h http.Header, cookie string, userAgent string) {
	setCookieHeader(h, cookie)
	h.Set("Referer", "https://www.fanbox.cc/")
	h.Set("User-Agent", userAgent)
	h.Set("Accept", "*/*")
	h.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	h.Set("Accept-Language", "en-US,en;q=0.5")
	h.Set("Priority", "u=4")
	h.Set("Sec-Fetch-Dest", "empty")
	h.Set("Sec-Fetch-Mode", "cors")
	h.Set("Sec-Fetch-Site", "same-site")
}

func setCookieHeader(h http.Header, cookie string) {
	if cookie != "" {
		h.Set("Cookie", cookie)
	}
}

func chromeSecCHUA(userAgent string) string {
	version := parseBrowserMajor(userAgent)
	if version == 0 {
		version = 133
	}
	major := fmt.Sprintf("%d", version)
	return fmt.Sprintf(`"Not(A:Brand";v="99", "Google Chrome";v="%s", "Chromium";v="%s"`, major, major)
}

var chromeAPIHeaderOrder = []string{
	"sec-ch-ua",
	"sec-ch-ua-mobile",
	"user-agent",
	"accept",
	"sec-ch-ua-platform",
	"origin",
	"sec-fetch-site",
	"sec-fetch-mode",
	"sec-fetch-dest",
	"referer",
	"accept-encoding",
	"accept-language",
	"cookie",
	"priority",
}

var chromeAssetHeaderOrder = []string{
	"sec-ch-ua",
	"sec-ch-ua-mobile",
	"user-agent",
	"sec-ch-ua-platform",
	"accept",
	"sec-fetch-site",
	"sec-fetch-mode",
	"sec-fetch-dest",
	"referer",
	"accept-encoding",
	"accept-language",
	"cookie",
	"priority",
}

var firefoxAPIHeaderOrder = []string{
	"user-agent",
	"accept",
	"accept-language",
	"accept-encoding",
	"origin",
	"referer",
	"sec-fetch-dest",
	"sec-fetch-mode",
	"sec-fetch-site",
	"cookie",
	"priority",
}

var firefoxAssetHeaderOrder = []string{
	"user-agent",
	"accept",
	"accept-language",
	"accept-encoding",
	"referer",
	"sec-fetch-dest",
	"sec-fetch-mode",
	"sec-fetch-site",
	"cookie",
	"priority",
}
