package app

import (
	"backend-go/constants"
	"backend-go/database/redisx"
	handlers "backend-go/internal/user/handler"
	repository "backend-go/internal/user/repository/mongoDb"
	redisRepository "backend-go/internal/user/repository/redis"
	"backend-go/internal/user/services"
	middleware "backend-go/middlewares"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

type App struct {
	DB            *mongo.Database
	redisDB       *redisx.Client
	UserMangoRepo repository.UserRepository
	UserRedisRepo redisRepository.UserRedisRepository
	UserService   services.UserService
	UserHandler   handlers.UserHandler
}

// NewApp initializes everything in one place
func NewApp(mongoDB *mongo.Database, redisDB *redisx.Client) (*App, error) {
	//Setup repositories,services,handlers
	mongoRepo := repository.NewUserRepository(mongoDB)
	redisRepo := redisRepository.NewUserCache(redisDB)

	service := services.NewUserService(mongoRepo, redisRepo)
	handler := handlers.NewUserHandler(service)

	return &App{
		DB:            mongoDB,
		redisDB:       redisDB,
		UserMangoRepo: mongoRepo,
		UserRedisRepo: redisRepo,
		UserService:   service,
		UserHandler:   handler,
	}, nil
}

func (a *App) RegisterRoutes(r *mux.Router) {
	// Apply limiter to the router with default config
	var defaultCfg = middleware.RateLimitConfig{
		RateLimit:       constants.GLOBAL_RATE_LIMITER_RATE,
		BurstLimit:      constants.GLOBAL_RATE_LIMITER_BURST,
		RemainingTokens: constants.GLOBAL_RATE_LIMITER_BURST - 1,
		TTL:             constants.GLOBAL_RATE_LIMITER_TTL,
		LastRefill:      time.Now(),
	}
	rl := middleware.NewRateLimiter(a.redisDB, defaultCfg)

	//specific route config
	rl.AddRouteLimit("/api/user/login", middleware.RateLimitConfig{
		RateLimit:       constants.LOGIN_RATE_LIMITER_RATE,
		BurstLimit:      constants.LOGIN_RATE_LIMITER_BURST,
		RemainingTokens: constants.LOGIN_RATE_LIMITER_BURST - 1,
		TTL:             constants.GLOBAL_RATE_LIMITER_TTL,
		LastRefill:      time.Now(),
	})
	rl.AddRouteLimit("/api/user/profile", middleware.RateLimitConfig{
		RateLimit:       constants.PROFILE_RATE_LIMITER_RATE,
		BurstLimit:      constants.PROFILE_RATE_LIMITER_BURST,
		RemainingTokens: constants.PROFILE_RATE_LIMITER_BURST - 1,
		TTL:             constants.GLOBAL_RATE_LIMITER_TTL,
		LastRefill:      time.Now(),
	})

	r.Use(rl.Limit)

	r.HandleFunc("/register", a.UserHandler.RegisterUser).Methods("POST")
	r.HandleFunc("/login", a.UserHandler.LoginUser).Methods("POST")
	r.Handle("/profile", middleware.AuthMiddleware(http.HandlerFunc(a.UserHandler.Profile), a.UserRedisRepo)).Methods("GET")
	r.Handle("/logout", middleware.AuthMiddleware(http.HandlerFunc(a.UserHandler.LogoutUser), a.UserRedisRepo)).Methods("POST")
	r.Handle("/access-token", middleware.RefreshAuthMiddleware(http.HandlerFunc(a.UserHandler.GetSilentAccesToken), a.UserRedisRepo)).Methods("GET")
}
