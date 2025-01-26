package repository

import (
	"github.com/google/uuid"
	"github.com/lukelittle/claroz/claroz-backend/internal/models"
)

type UserRepositoryInterface interface {
	Create(user *models.User) error
	GetByID(id uuid.UUID) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uuid.UUID) error
	FindByHandle(handle string) (*models.User, error)
	FindByDID(did string) (*models.User, error)
	GetRemoteUsers() ([]*models.User, error)
}
