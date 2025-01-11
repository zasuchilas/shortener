package trusted

import (
	"github.com/zasuchilas/shortener/internal/app/logger"
	"go.uber.org/zap"
	"net"
	"net/http"
)

// trustedSubnet contains props for middleware.
type trustedSubnet struct {
	ipNet *net.IPNet
}

// NewTrustedSubnet creates trustedSubnet struct.
func NewTrustedSubnet(cidr string) *trustedSubnet {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		logger.Log.Info("parsing CIDR of trusted subnet", zap.String("error", err.Error()))
	}
	return &trustedSubnet{ipNet: ipNet}
}

// Middleware implements trusted subnet middleware.
func (t *trustedSubnet) Middleware(h http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		if t.ipNet == nil {
			w.WriteHeader(http.StatusForbidden)
		}

		ip := net.ParseIP(r.Header.Get("X-Real-IP"))
		if ip == nil || !t.ipNet.Contains(ip) {
			w.WriteHeader(http.StatusForbidden)
		}

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}
