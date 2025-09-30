package userType

import model "backend-go/models"

type UserResponse struct {
	User         *model.User
	AccessToken  string
	RefreshToken string
}
