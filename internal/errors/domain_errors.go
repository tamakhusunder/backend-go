package domainerrors

import "errors"

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrSomethingWentWrong = errors.New("something wen wrong")

	ErrGeneratingJWTToken  = errors.New("error in generating JWT token")
	ErrStoringTokenInRedis = errors.New("error in redis store")
)
