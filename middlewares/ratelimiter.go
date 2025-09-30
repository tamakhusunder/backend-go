// internal/middleware/ratelimiter.go
package middleware

import (
	"backend-go/database/redisx"
	"backend-go/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenBucketState struct {
	RateLimit       int       `json:"rate_limit"`
	BurstLimit      int       `json:"burst_limit"`
	RemainingTokens int       `json:"remaining_tokens"`
	LastRefill      time.Time `json:"last_refill"`
}

type RateLimiter struct {
	redisClient *redisx.Client
	rate        int
	burst       int
}

func NewRateLimiter(redisClient *redisx.Client, rate, burst int) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		rate:        rate,
		burst:       burst,
	}
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := utils.GetClientIP(r)
		key := "ratelimit:" + ip
		ctx := context.Background()

		// Fetch from Redis
		val, err := rl.redisClient.Rdb.Get(ctx, key).Result()
		var state TokenBucketState

		log.Println("TokenBucketState:", val)

		if err == redis.Nil {
			// First time this IP
			state = TokenBucketState{
				RateLimit:       rl.rate,
				BurstLimit:      rl.burst,
				RemainingTokens: rl.burst - 1, // Consume one token immediately
				LastRefill:      time.Now(),
			}
			data, _ := json.Marshal(state)
			rl.redisClient.Rdb.Set(ctx, key, data, 0)
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Decode JSON state
		if err := json.Unmarshal([]byte(val), &state); err != nil {
			log.Print("Failed to unmarshal token bucket state:", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		log.Printf("RateLimiter State for %s: %+v\n", ip, state)

		// Refill tokens
		now := time.Now()
		elapsed := now.Sub(state.LastRefill).Minutes()
		newTokens := int(elapsed * float64(state.RateLimit))
		if newTokens > 0 {
			state.RemainingTokens = min(state.BurstLimit, state.RemainingTokens+newTokens)
			state.LastRefill = now
		}

		// Consume token
		if state.RemainingTokens > 0 {
			state.RemainingTokens--
			data, _ := json.Marshal(state)
			rl.redisClient.Rdb.Set(ctx, key, data, 0)
			next.ServeHTTP(w, r)
		} else {
			// http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "Too Many Requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			return
		}
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
