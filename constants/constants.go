package constants

import "time"

const ACCESS_TOKEN_EXPIRATION_IN_SECONDS = int64(30 * 60)           // 30 minutes in seconds
const REFRESH_TOKEN_EXPIRATION_IN_SECONDS = int64(3 * 24 * 60 * 60) // 3 day in seconds

const ACCESS_TOKEN_EXPIRATION time.Duration = time.Duration(ACCESS_TOKEN_EXPIRATION_IN_SECONDS) * time.Second   // 30 minutes
const REFRESH_TOKEN_EXPIRATION time.Duration = time.Duration(REFRESH_TOKEN_EXPIRATION_IN_SECONDS) * time.Second // 3 days

const RATE_LIMITER_RATE = 5                                            // tokens per minute
const RATE_LIMITER_BURST = 5                                           // max bucket size
const RATE_LIMITER_TTL time.Duration = time.Duration(10 * time.Minute) // key expiration time
