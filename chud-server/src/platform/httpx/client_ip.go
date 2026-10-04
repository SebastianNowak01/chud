package httpx

import (
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/sebnow/chud/platform/config"
)

func ClientIP(r *http.Request) string {
	if header := os.Getenv(config.RealIPHeader); header != "" {
		values := strings.Split(r.Header.Get(header), ",")
		if ip := strings.TrimSpace(values[len(values)-1]); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
