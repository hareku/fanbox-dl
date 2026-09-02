package tlsclient

import (
	"net/http"

	fhttp "github.com/bogdanfinn/fhttp"
)

// SetHeaderOrder stores fhttp's magic header-order keys on a net/http header.
func SetHeaderOrder(h http.Header, headerOrder, pseudoHeaderOrder []string) {
	if len(headerOrder) > 0 {
		h[fhttp.HeaderOrderKey] = append([]string(nil), headerOrder...)
	}
	if len(pseudoHeaderOrder) > 0 {
		h[fhttp.PHeaderOrderKey] = append([]string(nil), pseudoHeaderOrder...)
	}
}
