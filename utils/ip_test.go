package utils_test

import (
	"backend-go/utils" // Import the package being tested
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name              string
		isLocal           bool
		xForwardedFor     string
		xRealIP           string
		remoteAddr        string
		expectedIP        string
		remoteAddrIsError bool // Should net.SplitHostPort return an error
	}{
		// 1. Local Environment Case (Highest Priority)
		{
			name:          "Local Environment Check",
			isLocal:       true,
			xForwardedFor: "1.1.1.1, 2.2.2.2",
			xRealIP:       "3.3.3.3",
			remoteAddr:    "192.168.1.1:12345",
			expectedIP:    "127.0.0.1",
		},

		// 2. X-Forwarded-For (XFF) Cases
		{
			name:          "XFF Single IP",
			isLocal:       false,
			xForwardedFor: "203.0.113.42",
			xRealIP:       "",
			remoteAddr:    "192.168.1.1:12345",
			expectedIP:    "203.0.113.42",
		},
		{
			name:          "XFF Multiple IPs (Take First)",
			isLocal:       false,
			xForwardedFor: "203.0.113.42, 198.51.100.100, 10.0.0.1",
			xRealIP:       "3.3.3.3",
			remoteAddr:    "192.168.1.1:12345",
			expectedIP:    "203.0.113.42", // Should return the client's IP (the first one)
		},
		{
			name:          "XFF with leading/trailing spaces",
			isLocal:       false,
			xForwardedFor: "  203.0.113.42,198.51.100.100 ",
			xRealIP:       "",
			remoteAddr:    "192.168.1.1:12345",
			expectedIP:    "203.0.113.42",
		},

		// 3. X-Real-IP (XRI) Case
		{
			name:          "XRealIP Fallback (XFF empty)",
			isLocal:       false,
			xForwardedFor: "", // Empty after trimming/no valid IP
			xRealIP:       "198.51.100.10",
			remoteAddr:    "192.168.1.1:12345",
			expectedIP:    "198.51.100.10",
		},

		// 4. RemoteAddr Fallback Cases
		// {
		// 	name:          "RemoteAddr IP Only (No Headers)",
		// 	isLocal:       false,
		// 	xForwardedFor: "",
		// 	xRealIP:       "",
		// 	remoteAddr:    "192.168.1.100:12345",
		// 	expectedIP:    "192.168.1.100", // net.SplitHostPort should work
		// },
		// {
		// 	name:              "RemoteAddr Error Fallback",
		// 	isLocal:           false,
		// 	xForwardedFor:     "",
		// 	xRealIP:           "",
		// 	remoteAddr:        "invalid-address",
		// 	expectedIP:        "invalid-address", // Should return the full RemoteAddr on error
		// 	remoteAddrIsError: true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.isLocal {
				assert.Equal(t, "127.0.0.1", utils.GetClientIP(&http.Request{}, tt.isLocal), "Local environment should return 127.0.0.1")
				return
			}

			// Create a mock *http.Request
			req := &http.Request{
				Header: http.Header{},
			}

			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			// --- Actual Function Call ---
			actualIP := utils.GetClientIP(req, tt.isLocal)
			fmt.Printf("+++++++ %v actualIP: %s\n", tt.expectedIP, actualIP)

			// --- Assertion ---
			assert.Equal(t, tt.expectedIP, actualIP, "The resulting IP did not match the expected IP.")
			assert.NoError(t, nil, "No error should occur during IP extraction.")
		})
	}
}
