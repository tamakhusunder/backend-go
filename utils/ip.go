package utils

import (
	"backend-go/config"
	"log"
	"net"
	"net/http"
	"strings"
)

func GetClientIP(r *http.Request) string {
	if config.IsLocal() {
		return "127.0.0.1"
	}

	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				return ip
			}
		}
	}

	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("Failed to get client IP: %v", err)
		return r.RemoteAddr
	}

	//improvment : can add fallback value also
	return ip
}
