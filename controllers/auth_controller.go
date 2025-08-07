package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/dto"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/services"
	"github.com/minh6824pro/nxrGO/utils"
	"io"
	"net/http"
)

type AuthController struct {
	authService services.AuthService
}

func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{authService}
}

func (a *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if errors.Is(err, io.EOF) {
			customErr.WriteError(c, customErr.NewError(
				customErr.BAD_REQUEST,
				"Request body is empty",
				http.StatusBadRequest,
				err,
			))
			return
		}

		if utils.HandleValidationError(c, err) {
			return
		}
		customErr.WriteError(c, err)
		return
	}

	resp, err := a.authService.Register(req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (a *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if errors.Is(err, io.EOF) {
			customErr.WriteError(c, customErr.NewError(
				customErr.BAD_REQUEST,
				"Request body is empty",
				http.StatusBadRequest,
				err,
			))
			return
		}
		if utils.HandleValidationError(c, err) {
			return
		}
		customErr.WriteError(c, err)
		return
	}

	resp, err := a.authService.Login(req)
	if err != nil {
		customErr.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (a *AuthController) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		customErr.WriteError(c, customErr.NewError(
			customErr.UNAUTHORIZED,
			"Unauthorized",
			http.StatusUnauthorized,
			nil))
		return
	}

	profile, err := a.authService.GetProfile(userID.(uint))
	if err != nil {
		customErr.WriteError(c, customErr.NewError(
			customErr.UNAUTHORIZED,
			"User not found",
			http.StatusUnauthorized,
			nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": profile})
}

func (a *AuthController) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newAccessToken, err := a.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": newAccessToken})
}
