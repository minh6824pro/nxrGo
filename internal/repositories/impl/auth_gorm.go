package impl

import (
	"github.com/minh6824pro/nxrGO/internal/models"
	"github.com/minh6824pro/nxrGO/internal/repositories"
	"gorm.io/gorm"
)

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) repositories.AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *authRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *authRepository) IsEmailExists(email string) bool {
	var count int64
	r.db.Model(&models.User{}).Where("email = ?", email).Count(&count)
	return count > 0
}
