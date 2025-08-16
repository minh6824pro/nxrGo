package impl

import (
	"github.com/minh6824pro/nxrGO/dto"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/jwt"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"github.com/minh6824pro/nxrGO/services"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type authService struct {
	repo       repositories.AuthRepository
	jwtService *jwt.JWTService
}

func NewAuthService(repo repositories.AuthRepository, jwtSvc *jwt.JWTService) services.AuthService {
	return &authService{
		repo:       repo,
		jwtService: jwtSvc,
	}
}

func (s *authService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	if s.repo.IsEmailExists(req.Email) {

		return nil, customErr.NewError(customErr.DUPLICATE_ENTRY, "Email already exists", http.StatusBadRequest, nil)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Error generating password", http.StatusInternalServerError, err)
	}

	user := &models.User{
		FullName:    req.FullName,
		Email:       req.Email,
		Password:    string(hashed),
		PhoneNumber: req.PhoneNumber,
		Role:        models.RoleUser,
		Active:      1,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Error creating user", http.StatusInternalServerError, err)
	}

	accessToken, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Error generating token 1", http.StatusInternalServerError, err)
	}
	refreshToken, err := s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Error generating token 2", http.StatusInternalServerError, err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (s *authService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil || user.Active != uint8(1) {
		return nil, customErr.NewError(customErr.INVALID_CREDENTIALS, "Email not exists", http.StatusUnauthorized, nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, customErr.NewError(customErr.INVALID_CREDENTIALS, "Wrong password", http.StatusUnauthorized, nil)
	}

	now := time.Now()
	user.LastLogin = &now
	_ = s.repo.Update(user)

	accessToken, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (s *authService) GetProfile(userID uint) (*models.User, error) {
	return s.repo.FindByID(userID)
}

func (s *authService) RefreshToken(refreshToken string) (string, error) {
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	user, err := s.repo.FindByID(claims.UserID)
	if err != nil {
		return "", customErr.NewError(customErr.ITEM_NOT_FOUND, "User not found", http.StatusUnauthorized, nil)
	}

	return s.jwtService.GenerateToken(user)
}
