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

	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, creds model.User) (interface{}, error)
	Login(ctx context.Context, email string, password string, clientIp string) (*userType.UserResponse, error)
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

	// store token in redis
	if _, storeErr := s.redisRepo.StoreToken(ctx, user.ID, refreshToken, clientIp); storeErr != nil {
		fmt.Print("Error storing token in redis", storeErr)
		return nil, domainerrors.ErrStoringTokenInRedis
	}
	resp := &userType.UserResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return resp, nil
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
	refreshToken, errRefreshToken := utils.GenerateRefreshToken(userId, email)

	if errAcessToken != nil || errRefreshToken != nil {
		fmt.Print("Error generating JWT token", errAcessToken)
		return "", domainerrors.ErrGeneratingJWTToken
	}

	// store token in redis
	if _, storeErr := s.redisRepo.StoreToken(ctx, userId, refreshToken, clientIp); storeErr != nil {
		fmt.Print("Error storing token in redis", storeErr)
		return "", domainerrors.ErrStoringTokenInRedis
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
