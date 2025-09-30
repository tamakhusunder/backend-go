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
	// Apply limiter to the router
	rl := middleware.NewRateLimiter(a.redisDB, constants.RATE_LIMITER_RATE, constants.RATE_LIMITER_BURST)
	r.Use(rl.Limit)

	r.Handle("", middleware.AuthMiddleware(http.HandlerFunc(a.UserHandler.Profile))).Methods("GET")
	r.HandleFunc("/register", a.UserHandler.RegisterUser).Methods("POST")
	r.HandleFunc("/login", a.UserHandler.LoginUser).Methods("POST")
	r.Handle("/logout", middleware.AuthMiddleware(http.HandlerFunc(a.UserHandler.LogoutUser))).Methods("POST")
	r.Handle("/access-token", middleware.AuthMiddleware(http.HandlerFunc(a.UserHandler.GetSilentAccesToken))).Methods("GET")
}
