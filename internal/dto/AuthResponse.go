package dto

import "github.com/minh6824pro/nxrGO/internal/models"

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         *models.User `json:"user"`
}
