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
	DeleteToken(ctx context.Context, userID string) (interface{}, error)
	SetBlacklistOfAccessToken(ctx context.Context, userID string, accessToken string, ttlTime time.Duration) (interface{}, error)
	IsBlacklistedAccessToken(ctx context.Context, userID string, accessToken string) (bool, error)
}

type userCacheImpl struct {
	redis *redisx.Client
}

func NewUserCache(rDb *redisx.Client) UserRedisRepository {
	return &userCacheImpl{
		redis: rDb,
	}
}

// methods for storing and deleting refresh tokens(used during login and logout)
func (r *userCacheImpl) StoreToken(ctx context.Context, userID string, refreshToken string, clientIp string) (interface{}, error) {
	if rErr := redisx.Rdb.HSet(ctx, userID, map[string]interface{}{
		"refreshToken": refreshToken,
		"ipAddress":    clientIp,
	}).Err(); rErr != nil {
		log.Printf("Failed to set user session in Redis: %v", rErr)
		return nil, rErr
	}

	return nil, nil
}

func (r *userCacheImpl) DeleteToken(ctx context.Context, userID string) (interface{}, error) {
	exists, err := redisx.Rdb.Exists(ctx, userID).Result()
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
