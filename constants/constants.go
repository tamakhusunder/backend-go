package constants

import (
	"time"
)

// JWT tokens
const ACCESS_TOKEN_EXPIRATION_IN_SECONDS = int64(30 * 60)           // 30 minutes in seconds
const REFRESH_TOKEN_EXPIRATION_IN_SECONDS = int64(3 * 24 * 60 * 60) // 3 day in seconds

const ACCESS_TOKEN_EXPIRATION time.Duration = time.Duration(ACCESS_TOKEN_EXPIRATION_IN_SECONDS) * time.Second   // 30 minutes
const REFRESH_TOKEN_EXPIRATION time.Duration = time.Duration(REFRESH_TOKEN_EXPIRATION_IN_SECONDS) * time.Second // 3 days

// Rate Limiter settings
const GLOBAL_RATE_LIMITER_RATE = 5                                          // in minutes                                     // tokens per minute
const GLOBAL_RATE_LIMITER_BURST = 5                                         // max bucket size
const GLOBAL_RATE_LIMITER_TTL time.Duration = time.Duration(24 * time.Hour) // key expiration time
const INFINITY_TTL time.Duration = 0

const LOGIN_RATE_LIMITER_RATE = 1  // tokens per minute
const LOGIN_RATE_LIMITER_BURST = 5 // max bucket size

const ME_RATE_LIMITER_RATE = 10  // tokens per minute
const ME_RATE_LIMITER_BURST = 10 // max bucket size

// Blacklist settings
const BLACKLIST_ACCESS_TOKEN string = "blacklistAcessToken"
