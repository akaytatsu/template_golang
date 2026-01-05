package repository

import (
	"app/entity"

	"gorm.io/gorm"
)

type RepositoryUser struct {
	DB *gorm.DB
}

func NewUserPostgres(DB *gorm.DB) *RepositoryUser {
	return &RepositoryUser{DB: DB}
}

func (u *RepositoryUser) GetByID(id int) (user *entity.EntityUser, err error) {
	err = u.DB.First(&user, id).Error
	return user, err
}

func (u *RepositoryUser) GetByMail(email string) (user *entity.EntityUser, err error) {
	err = u.DB.Where("email = ?", email).First(&user).Error
	return user, err
}

func (u *RepositoryUser) CreateUser(user *entity.EntityUser) error {
	return u.DB.Create(&user).Error
}

func (u *RepositoryUser) UpdateUser(user *entity.EntityUser) error {
	// Verify user exists before updating
	var existingUser entity.EntityUser
	err := u.DB.Where("email = ?", user.Email).First(&existingUser).Error
	if err != nil {
		return err
	}

	return u.DB.Save(&user).Error
}

func (u *RepositoryUser) DeleteUser(user *entity.EntityUser) error {
	// Verify user exists before deleting
	var existingUser entity.EntityUser
	err := u.DB.Where("email = ?", user.Email).First(&existingUser).Error
	if err != nil {
		return err
	}

	return u.DB.Delete(&user).Error
}

func (u *RepositoryUser) GetUsersFromIDs(ids []int) (users []entity.EntityUser, err error) {
	users = make([]entity.EntityUser, 0)
	err = u.DB.Where("id IN ?", ids).Find(&users).Error
	return users, err
}

func (u *RepositoryUser) GetUsers(filters entity.EntityUserFilters) (users []entity.EntityUser, err error) {
	users = make([]entity.EntityUser, 0)

	query := u.DB

	if filters.Search != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	// Handle Active filter - convert string to boolean
	if filters.Active != "" {
		if filters.Active == "true" {
			query = query.Where("active = ?", true)
		} else if filters.Active == "false" {
			query = query.Where("active = ?", false)
		}
	}

	err = query.Find(&users).Error
	return users, err
}

// GetUser is an alias for GetByID - kept for compatibility
func (u *RepositoryUser) GetUser(id int) (user *entity.EntityUser, err error) {
	return u.GetByID(id)
}
