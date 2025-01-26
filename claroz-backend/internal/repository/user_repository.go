package repository

import (
	"github.com/google/uuid"
	"github.com/lukelittle/claroz/claroz-backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

func (r *UserRepository) FindByHandle(handle string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "handle = ?", handle).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByDID(did string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "d_id = ?", did).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetRemoteUsers() ([]*models.User, error) {
	var users []*models.User
	err := r.db.Where("federation_type = ?", "remote").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
