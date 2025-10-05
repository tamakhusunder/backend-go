package rdsModel

import "time"

type RateLimitConfig struct {
	RateLimit       int           `json:"rate_limit"`
	BurstLimit      int           `json:"burst_limit"`
	RemainingTokens int           `json:"remaining_tokens"`
	TTL             time.Duration `json:"ttl"`
	LastRefill      time.Time     `json:"last_refill"`
}
