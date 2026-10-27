package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestClientIPResolution guards how the real router resolves c.ClientIP(), which
// the analytics middleware records for every public request. In production every request
// arrives from nginx on loopback with Cloudflare's CF-Connecting-IP; the client
// fully controls X-Forwarded-For and X-Real-IP.
func TestClientIPResolution(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := setupRouter()
	// echo the resolved address; registered after setupRouter so the request
	// still runs through the global middleware
	router.GET("/test/clientip", func(c *gin.Context) {
		c.String(http.StatusOK, c.ClientIP())
	})

	tests := []struct {
		name       string
		remoteAddr string
		header     map[string]string
		wantIP     string
	}{
		{
			name:       "client ip comes from CF-Connecting-IP",
			remoteAddr: "127.0.0.1:40000",
			header:     map[string]string{"CF-Connecting-IP": "198.51.100.1", "X-Forwarded-For": "198.51.100.1"},
			wantIP:     "198.51.100.1",
		},
		{
			name:       "forged X-Forwarded-For is ignored",
			remoteAddr: "127.0.0.1:40000",
			header:     map[string]string{"CF-Connecting-IP": "198.51.100.2", "X-Forwarded-For": "203.0.113.9, 198.51.100.2"},
			wantIP:     "198.51.100.2",
		},
		{
			name:       "unparseable X-Forwarded-For does not fall back to loopback",
			remoteAddr: "127.0.0.1:40000",
			header:     map[string]string{"CF-Connecting-IP": "198.51.100.3", "X-Forwarded-For": "notanip"},
			wantIP:     "198.51.100.3",
		},
		{
			name:       "forged X-Real-IP is ignored",
			remoteAddr: "127.0.0.1:40000",
			header:     map[string]string{"CF-Connecting-IP": "198.51.100.4", "X-Real-IP": "203.0.113.9"},
			wantIP:     "198.51.100.4",
		},
		{
			name:       "ipv6 client",
			remoteAddr: "127.0.0.1:40000",
			header:     map[string]string{"CF-Connecting-IP": "2001:db8::5"},
			wantIP:     "2001:db8::5",
		},
		{
			name:       "nginx over ipv6 loopback is trusted",
			remoteAddr: "[::1]:40000",
			header:     map[string]string{"CF-Connecting-IP": "198.51.100.6"},
			wantIP:     "198.51.100.6",
		},
		{
			name:       "CF-Connecting-IP from a non-loopback peer is ignored",
			remoteAddr: "192.0.2.7:40000",
			header:     map[string]string{"CF-Connecting-IP": "198.51.100.7", "X-Forwarded-For": "198.51.100.7"},
			wantIP:     "192.0.2.7",
		},
		{
			name:       "forwarding headers without CF-Connecting-IP are ignored",
			remoteAddr: "127.0.0.1:40000",
			header:     map[string]string{"X-Forwarded-For": "203.0.113.9", "X-Real-IP": "203.0.113.9"},
			wantIP:     "127.0.0.1",
		},
		{
			name:       "unparseable CF-Connecting-IP falls back to the peer",
			remoteAddr: "127.0.0.1:40000",
			header:     map[string]string{"CF-Connecting-IP": "notanip", "X-Forwarded-For": "203.0.113.9"},
			wantIP:     "127.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test/clientip", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.header {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.wantIP, w.Body.String(), "resolved client ip")
		})
	}
}
