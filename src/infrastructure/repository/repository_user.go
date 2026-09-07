package repository

import (
	"app/entity"

	"gorm.io/gorm"
)

type RepositoryUser struct {
	DB *gorm.DB
}

func NewUserPostgres(db *gorm.DB) *RepositoryUser {
	return &RepositoryUser{DB: db}
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
	_, err := u.GetByMail(user.Email)
	if err != nil {
		return err
	}

	return u.DB.Save(&user).Error
}

func (u *RepositoryUser) DeleteUser(user *entity.EntityUser) error {
	_, err := u.GetByMail(user.Email)
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

// applyUserFilters aplica os filtros de busca sem paginação, para que a mesma
// cláusula sirva ao Count e ao Find.
func (u *RepositoryUser) applyUserFilters(filters entity.EntityUserFilters) *gorm.DB {
	query := u.DB.Model(&entity.EntityUser{})

	if filters.Search != "" {
		query = query.Where("name LIKE ? or email LIKE ?", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	if filters.Active != "" {
		query = query.Where("active = ?", filters.Active)
	}

	if len(filters.IDs) > 0 {
		query = query.Where("id IN ?", filters.IDs)
	}

	return query
}

func (u *RepositoryUser) GetUsers(filters entity.EntityUserFilters) (users []entity.EntityUser, total int64, err error) {
	users = make([]entity.EntityUser, 0)

	if err = u.applyUserFilters(filters).Count(&total).Error; err != nil {
		return users, 0, err
	}

	query := u.applyUserFilters(filters)

	if filters.PageSize > 0 {
		query = query.Limit(filters.PageSize).Offset(filters.Page * filters.PageSize)
	}

	err = query.Find(&users).Error

	return users, total, err
}

func (u *RepositoryUser) GetUser(id int) (user *entity.EntityUser, err error) {
	err = u.DB.First(&user, id).Error

	return user, err
}
