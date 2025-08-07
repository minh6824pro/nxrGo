package services

import (
	"github.com/minh6824pro/nxrGO/dto"
	"github.com/minh6824pro/nxrGO/models"
)

type AuthService interface {
	Register(dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(dto.LoginRequest) (*dto.AuthResponse, error)
	GetProfile(userID uint) (*models.User, error)
	RefreshToken(refreshToken string) (string, error)
}
