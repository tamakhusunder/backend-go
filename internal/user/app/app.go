package app

import (
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
		UserMangoRepo: mongoRepo,
		UserRedisRepo: redisRepo,
		UserService:   service,
		UserHandler:   handler,
	}, nil
}

func (a *App) RegisterRoutes(r *mux.Router) {
	r.Handle("", middleware.AuthMiddleware(http.HandlerFunc(a.UserHandler.Profile))).Methods("GET")
	r.HandleFunc("/register", a.UserHandler.RegisterUser).Methods("POST")
	r.HandleFunc("/login", a.UserHandler.LoginUser).Methods("POST")
	r.Handle("/logout", middleware.AuthMiddleware(http.HandlerFunc(a.UserHandler.LogoutUser))).Methods("POST")
	r.Handle("/access-token", middleware.AuthMiddleware(http.HandlerFunc(a.UserHandler.GetSilentAccesToken))).Methods("GET")
}
