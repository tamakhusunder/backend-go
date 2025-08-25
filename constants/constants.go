package constants

import "time"

const ACCESS_TOKEN_EXPIRATION_IN_SECONDS = int64(30 * 60)           // 30 minutes in seconds
const REFRESH_TOKEN_EXPIRATION_IN_SECONDS = int64(3 * 24 * 60 * 60) // 3 day in seconds

const ACCESS_TOKEN_EXPIRATION time.Duration = time.Duration(ACCESS_TOKEN_EXPIRATION_IN_SECONDS) * time.Second   // 30 minutes
const REFRESH_TOKEN_EXPIRATION time.Duration = time.Duration(REFRESH_TOKEN_EXPIRATION_IN_SECONDS) * time.Second // 3 days
