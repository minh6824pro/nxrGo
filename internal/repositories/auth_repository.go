package repositories

import (
	"github.com/minh6824pro/nxrGO/internal/models"
)

type AuthRepository interface {
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	IsEmailExists(email string) bool
}
