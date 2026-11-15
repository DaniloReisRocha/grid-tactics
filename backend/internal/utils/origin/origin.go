package origin

import (
	"net/url"
	"strings"
)

// Allowed accepts exact origins and URL patterns such as https://*.ngrok-free.app.
// The wildcard matches subdomains only; schemes and ports must still match.
func Allowed(value string, patterns []string) bool {
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return false
	}
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if value == pattern {
			return true
		}
		p, err := url.Parse(pattern)
		if err != nil || p.Scheme != u.Scheme || p.Port() != u.Port() || p.User != nil || p.Path != "" || p.RawQuery != "" || p.Fragment != "" {
			continue
		}
		host := strings.ToLower(p.Hostname())
		if strings.HasPrefix(host, "*.") && strings.Count(host, "*") == 1 {
			suffix := host[1:]
			candidate := strings.ToLower(u.Hostname())
			if strings.HasSuffix(candidate, suffix) && len(candidate) > len(suffix) {
				return true
			}
		}
	}
	return false
}
