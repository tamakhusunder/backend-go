package domainerrors

import "errors"

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrSomethingWentWrong = errors.New("something went wrong")

	ErrGeneratingJWTToken  = errors.New("error in generating JWT token")
	ErrStoringTokenInRedis = errors.New("error in redis store")
	ErrStoringTokenInDb    = errors.New("error in Database")

	ErrCacheMiss = errors.New("cache miss")
)
