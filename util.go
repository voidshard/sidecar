package main

import (
	"net/http"
	"strings"
)

// determineIP determines the IP address of the client from the request.
//
// We'll check an array of possible headers and choose the first non-empty one.
// If all else fails we'll use the RemoteAddr.
func determineIP(r *http.Request) (map[string]interface{}, string) {
	attrs := map[string]interface{}{
		"addr.remote_addr": r.RemoteAddr,
		"http.user_agent":  r.UserAgent(),
		"http.referer":     r.Referer(),
		"addr.chosen":      "r.RemoteAddr",
	}
	chosen := ""
	for _, header := range []string{
		"CF-Pseudo-IPv4",
		"CF-Connecting-IP",
		"CF-Connecting-IPv6",
		"X-Forwarded-For",
		"X-Real-IP",
	} {
		ip := r.Header.Get(header)
		if chosen == "" && ip != "" {
			// take the first non-empty IP address, we'll keep the others as attrs
			chosen = ip
			attrs["addr.chosen"] = header
		}
		attrs["addr."+strings.ToLower(header)] = ip
	}
	if chosen == "" {
		chosen = r.RemoteAddr
	}
	return attrs, chosen
}
