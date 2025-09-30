package repository

import (
	"backend-go/database/redisx"
	"context"
	"log"
)

type UserRedisRepository interface {
	StoreToken(ctx context.Context, userID string, refreshToken string, clientIp string) (interface{}, error)
	DeleteToken(ctx context.Context, userID string) (interface{}, error)
}

type userCacheImpl struct {
	redis *redisx.Client
}

func NewUserCache(rDb *redisx.Client) UserRedisRepository {
	return &userCacheImpl{
		redis: rDb,
	}
}

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
