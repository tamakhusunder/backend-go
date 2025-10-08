package services

import (
	"backend-go/constants"
	domainerrors "backend-go/constants/errors"
	repository "backend-go/internal/user/repository/mongoDb"
	redisRepository "backend-go/internal/user/repository/redis"
	model "backend-go/models"
	userType "backend-go/type"
	"backend-go/utils"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, creds model.User) (interface{}, error)
	Login(ctx context.Context, email string, password string, clientIp string) (*userType.UserResponse, error)
	Profile(ctx context.Context, UserId string, clientIp string) (*model.User, error)
	Logout(ctx context.Context, userId string, accessToken string) (interface{}, error)
	GetSilentAccessToken(ctx context.Context, userId string, email string, clientIp string) (string, error)
}

type UserServiceImpl struct {
	repo      repository.UserRepository
	redisRepo redisRepository.UserRedisRepository
}

func NewUserService(r repository.UserRepository, redisRepo redisRepository.UserRedisRepository) *UserServiceImpl {
	return &UserServiceImpl{
		repo:      r,
		redisRepo: redisRepo,
	}
}

func (s *UserServiceImpl) Register(ctx context.Context, creds model.User) (interface{}, error) {
	res, err := s.repo.Create(ctx, creds)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return nil, err
	}

	return res, nil
}

func (s *UserServiceImpl) Login(ctx context.Context, email string, password string, clientIp string) (*userType.UserResponse, error) {
	//check if user exists
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		log.Printf("userService_Login:Failed to fetch user from database: %v", err)
		if err == mongo.ErrNoDocuments {
			return nil, domainerrors.ErrUserNotFound
		}
		return nil, err
	}
	if user == nil {
		return nil, domainerrors.ErrUserNotFound
	}

	//compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, domainerrors.ErrInvalidCredentials
	}

	// Generate JWT token
	accessToken, errAcessToken := utils.GenerateAccessToken(user.ID, user.Email)
	refreshToken, errRefreshToken := utils.GenerateRefreshToken(user.ID, user.Email)

	if errAcessToken != nil || errRefreshToken != nil {
		fmt.Print("Error generating JWT token", errAcessToken, errRefreshToken)
		return nil, domainerrors.ErrGeneratingJWTToken
	}

	// store token in redis and mongodb
	if _, storeErr := s.redisRepo.StoreToken(ctx, user.ID, refreshToken, clientIp); storeErr != nil {
		fmt.Print("Error storing token in redis", storeErr)
		return nil, domainerrors.ErrStoringTokenInRedis
	}

	updates := bson.M{
		"token":      refreshToken,
		"ip_address": clientIp,
	}
	if _, dbErr := s.updateUserByID(ctx, user.ID, updates); dbErr != nil {
		fmt.Print("Error storing token in Database", dbErr)
		return nil, domainerrors.ErrStoringTokenInDb
	}

	resp := &userType.UserResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return resp, nil
}

// cache aside stategy for user profile
func (s *UserServiceImpl) Profile(ctx context.Context, UserId string, clientIp string) (*model.User, error) {
	//first check in redis cache
	cachedUser, cachedErr := s.redisRepo.GetUser(ctx, UserId)
	if cachedErr != nil || cachedErr == redis.Nil || cachedUser == nil {
		// cache miss, fetch from db
		user, err := s.repo.FindByID(ctx, UserId)
		if err != nil {
			log.Printf("userService.Profile: Failed to fetch user from database after cache hit: %v", err)
			if err == mongo.ErrNoDocuments {
				return nil, domainerrors.ErrUserNotFound
			}
			return nil, err
		}
		if user == nil {
			log.Printf("userService.Profile: User not found in database after cache hit: %v", UserId)
			return nil, domainerrors.ErrUserNotFound
		}
		log.Println("User cache miss and fetched from database", cachedUser)

		_, saveErr := s.redisRepo.SaveUser(ctx, *user)
		if saveErr != nil {
			log.Printf("Failed to save user profile in Redis: %v", saveErr)
			return nil, saveErr
		}
		log.Printf("userService.Profile: User profile saved in Redis: %s", user.ID)

		return user, nil
	}

	return cachedUser, nil
}

func (s *UserServiceImpl) Logout(ctx context.Context, userId string, accessToken string) (interface{}, error) {
	_, err := s.redisRepo.DeleteToken(ctx, userId)
	if err != nil {
		return nil, err
	}

	_, err = s.redisRepo.SetBlacklistOfAccessToken(ctx, userId, accessToken, constants.ACCESS_TOKEN_EXPIRATION)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
func (s *UserServiceImpl) GetSilentAccessToken(ctx context.Context, userId string, email string, clientIp string) (string, error) {
	accessToken, errAcessToken := utils.GenerateAccessToken(userId, email)
	if errAcessToken != nil {
		fmt.Print("Error generating JWT token", errAcessToken)
		return "", domainerrors.ErrGeneratingJWTToken
	}

	return accessToken, nil
}

func (s *UserServiceImpl) FindByEmail(ctx context.Context, email string) (interface{}, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *UserServiceImpl) FindByID(ctx context.Context, id string) (interface{}, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *UserServiceImpl) updateUserByID(ctx context.Context, userId string, updatedData bson.M) (interface{}, error) {
	updatedUser, err := s.repo.UpdateByID(ctx, userId, updatedData)
	if err != nil {
		log.Printf("Failed to update user: %v", err)
		return nil, err
	}
	if updatedUser == nil {
		log.Printf("User not found for update: %v", userId)
		return nil, errors.New("user not found")
	}

	// clear cache of user while updating user profile (cache aside strategy- clear on write)
	_, clearErr := s.redisRepo.DeleteUser(ctx, updatedUser.ID)
	if clearErr != nil {
		log.Printf("Failed to delete user profile in Redis: %v", clearErr)
	}

	return updatedUser, nil
}
