package utils

import "net"

func GetRealIP(req *http.Request) string {
	// Try CloudFlare header
	if ip := req.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	return ip
}
