package repository

import (
	"app/entity"

	"gorm.io/gorm"
)

type RepositoryUser struct {
	db *gorm.DB
}

func NewUserPostgres(db *gorm.DB) *RepositoryUser {
	return &RepositoryUser{db: db}
}

func (u *RepositoryUser) GetByMail(email string) (user *entity.EntityUser, err error) {
	err = u.db.Where("email = ?", email).First(&user).Error

	return user, err
}

func (u *RepositoryUser) CreateUser(user *entity.EntityUser) error {
	return u.db.Create(&user).Error
}

func (u *RepositoryUser) UpdateUser(user *entity.EntityUser) error {

	_, err := u.GetByMail(user.Email)

	if err != nil {
		return err
	}

	return u.db.Save(&user).Error
}

func (u *RepositoryUser) DeleteUser(user *entity.EntityUser) error {

	_, err := u.GetByMail(user.Email)

	if err != nil {
		return err
	}

	return u.db.Delete(&user).Error
}
