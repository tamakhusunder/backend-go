package userType

import (
	model "backend-go/models"
	"backend-go/utils"
)

type UserResponse struct {
	User         *model.User
	AccessToken  string
	RefreshToken string
}

type UserContents struct {
	Claims      *utils.Claims
	AccessToken string
}
