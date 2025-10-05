package middleware

import (
	"backend-go/database/redisx"
	rdsModel "backend-go/models/redis"
	"backend-go/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redisClient *redisx.Client
	defaultCfg  rdsModel.RateLimitConfig
	routeCfg    map[string]rdsModel.RateLimitConfig
}

func NewRateLimiter(redisClient *redisx.Client, defaultCfg rdsModel.RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		defaultCfg:  defaultCfg,
		routeCfg:    make(map[string]rdsModel.RateLimitConfig),
	}
}

func (rl *RateLimiter) AddRouteLimit(path string, cfg rdsModel.RateLimitConfig) {
	rl.routeCfg[path] = cfg
}

// Token Bucket Algorithm
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := utils.GetClientIP(r)
		path := r.URL.Path
		key := "ratelimit:" + ip + ":" + path
		ctx := context.Background()

		cfg, exists := rl.routeCfg[path] // Check if specific config exists for the route
		if !exists {
			cfg = rl.defaultCfg
		}

		// Fetch from Redis
		val, err := rl.redisClient.Rdb.Get(ctx, key).Result()
		log.Println("TokenBucketState:", val)

		if err == redis.Nil {
			// First time this IP
			firstCfg := rdsModel.RateLimitConfig{
				RateLimit:       cfg.RateLimit,
				BurstLimit:      cfg.BurstLimit,
				RemainingTokens: cfg.RemainingTokens,
				TTL:             cfg.TTL,
				LastRefill:      cfg.LastRefill,
			}
			data, _ := json.Marshal(firstCfg)
			rl.redisClient.Rdb.Set(ctx, key, data, cfg.TTL)
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Decode JSON state from redis
		if err := json.Unmarshal([]byte(val), &cfg); err != nil {
			log.Print("Failed to unmarshal token bucket state:", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		log.Printf("RateLimiter State for IP %s and Path %s: %+v\n", ip, path, cfg)

		rl.refill(&cfg)

		// Consume token
		if cfg.RemainingTokens > 0 {
			cfg.RemainingTokens--
			data, _ := json.Marshal(cfg)
			rl.redisClient.Rdb.Set(ctx, key, data, cfg.TTL)
			next.ServeHTTP(w, r)
		} else {
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

// Refill tokens per minute
func (rl *RateLimiter) refill(cfg *rdsModel.RateLimitConfig) {
	now := time.Now()
	elapsed := now.Sub(cfg.LastRefill).Minutes()
	newTokens := int(elapsed * float64(cfg.RateLimit))
	if newTokens > 0 {
		cfg.RemainingTokens = utils.Min(cfg.BurstLimit, cfg.RemainingTokens+newTokens)
		cfg.LastRefill = now
	}
}
