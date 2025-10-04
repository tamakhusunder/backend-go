package repository

import (
	"backend-go/constants"
	"backend-go/database/redisx"
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type UserRedisRepository interface {
	StoreToken(ctx context.Context, userID string, refreshToken string, clientIp string) (interface{}, error)
	GetToken(ctx context.Context, userID string) (*RdsTokenSession, error)
	DeleteToken(ctx context.Context, userID string) (interface{}, error)
	SetBlacklistOfAccessToken(ctx context.Context, userID string, accessToken string, ttlTime time.Duration) (interface{}, error)
	IsBlacklistedAccessToken(ctx context.Context, userID string, accessToken string) (bool, error)
}

type userCacheImpl struct {
	redis *redisx.Client
}

type RdsTokenSession struct {
	RefreshToken string `redis:"refreshToken"`
	IPAddress    string `redis:"ipAddress"`
}

func NewUserCache(rDb *redisx.Client) UserRedisRepository {
	return &userCacheImpl{
		redis: rDb,
	}
}

// methods for storing and deleting refresh tokens(used during login and logout)
func (r *userCacheImpl) StoreToken(ctx context.Context, userID string, refreshToken string, clientIp string) (interface{}, error) {
	key := "refreshTokenWithIp:" + userID
	sessionData := RdsTokenSession{
		RefreshToken: refreshToken,
		IPAddress:    clientIp,
	}
	if rErr := redisx.Rdb.HSet(ctx, key, sessionData).Err(); rErr != nil {
		log.Printf("Failed to set user session in Redis: %v", rErr)
		return nil, rErr
	}

	if rErr := redisx.Rdb.Expire(ctx, key, constants.REFRESH_TOKEN_EXPIRATION).Err(); rErr != nil {
		log.Printf("Failed to set expiration for user session in Redis: %v", rErr)
		return nil, rErr
	}

	return nil, nil
}

func (r *userCacheImpl) GetToken(ctx context.Context, userID string) (*RdsTokenSession, error) {
	key := "refreshTokenWithIp:" + userID
	session := &RdsTokenSession{}

	err := redisx.Rdb.HGetAll(ctx, key).Scan(session)
	if err == redis.Nil {
		log.Printf("Key not found in redis: %v", err)
		return nil, err
	}
	if err != nil {
		log.Printf("Failed to retrieve or scan session data: %v", err)
		return nil, nil // No session found
	}

	return session, nil
}

func (r *userCacheImpl) DeleteToken(ctx context.Context, userID string) (interface{}, error) {
	key := "refreshTokenWithIp:" + userID
	exists, err := redisx.Rdb.Exists(ctx, key).Result()
	if err != nil {
		log.Printf("Failed to delete user session in Redis: %v", err)
		return nil, err
	}

	if exists == 0 {
		return nil, nil // userID does not exist in Redis
	}

	if rErr := redisx.Rdb.Del(ctx, userID).Err(); rErr != nil {
		log.Printf("Failed to delete user session in Redis: %v", rErr)
		return nil, rErr
	}
	log.Printf("User session deleted from Redis: %s", userID)

	return nil, nil
}

// methods for blacklisting access tokens (used during logout)
func (r *userCacheImpl) SetBlacklistOfAccessToken(ctx context.Context, userID string, accessTokentoken string, ttlTime time.Duration) (interface{}, error) {
	key := constants.BLACKLIST_ACCESS_TOKEN + ":" + userID

	exists, _ := redisx.Rdb.Exists(ctx, key).Result()
	if exists > 0 {
		return nil, nil // Token is already blacklisted
	}

	if rErr := redisx.Rdb.Set(ctx, key, accessTokentoken, ttlTime).Err(); rErr != nil {
		log.Printf("Failed to set blacklisted access token in Redis: %v", rErr)
		return nil, rErr
	}

	return nil, nil
}

func (r *userCacheImpl) IsBlacklistedAccessToken(ctx context.Context, userId string, accessTokentoken string) (bool, error) {
	key := constants.BLACKLIST_ACCESS_TOKEN + ":" + userId
	res, err := redisx.Rdb.Get(ctx, key).Result()
	if redis.Nil == err {
		return false, nil
	}
	if err != nil {
		log.Printf("Failed to check if access token is blacklisted in Redis: %v", err)
		return false, err
	}

	return res == accessTokentoken, nil
}
