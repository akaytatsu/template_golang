package usecase_user

import (
	"app/entity"
	"errors"
)

type UseCaseUser struct {
	repo IRepositoryUser
}

func NewService(repository IRepositoryUser) *UseCaseUser {
	return &UseCaseUser{repo: repository}
}

func (u *UseCaseUser) LoginUser(email string, password string) (*entity.EntityUser, error) {
	user, err := u.repo.GetByMail(email)
	if err != nil {
		return nil, err
	}

	err = user.ValidatePassword(password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UseCaseUser) Create(user *entity.EntityUser) error {
	err := user.GetValidated()
	if err != nil {
		return err
	}

	return u.repo.CreateUser(user)
}

func (u *UseCaseUser) Update(user *entity.EntityUser) error {
	return u.repo.UpdateUser(user)
}

func (u *UseCaseUser) Delete(user *entity.EntityUser) error {
	return u.repo.DeleteUser(user)
}

func (u *UseCaseUser) GetUserByToken(token string) (*entity.EntityUser, error) {
	claims, err := (&entity.EntityUser{}).ValidateToken(token)
	if err != nil {
		return nil, err
	}

	user, err := u.repo.GetByID(claims.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UseCaseUser) UpdatePassword(id int, oldPassword, newPassword, confirmPassword string) error {
	user, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}

	err = user.ValidatePassword(oldPassword)
	if err != nil {
		return err
	}

	if newPassword != confirmPassword {
		return errors.New("passwords do not match")
	}

	user.UpdatePassword(newPassword)

	err = user.GetValidated()
	if err != nil {
		return err
	}

	err = u.repo.UpdateUser(user)

	return err
}

func (u *UseCaseUser) GetUsersFromIDs(ids []int) (users []entity.EntityUser, err error) {
	return u.repo.GetUsersFromIDs(ids)
}

func (u *UseCaseUser) GetUsers(filters entity.EntityUserFilters) (users []entity.EntityUser, err error) {
	return u.repo.GetUsers(filters)
}

func (u *UseCaseUser) GetUser(id int) (user *entity.EntityUser, err error) {
	return u.repo.GetUser(id)
}

func JWTTokenGenerator(u entity.EntityUser) (signedToken string, signedRefreshToken string, err error) {
	return u.JWTTokenGenerator()
}
