package fanbox

import (
	"net/http"
	"net/http/cookiejar"
)

func NewCookieJar() http.CookieJar {
	jar, _ := cookiejar.New(nil)
	return jar
}
