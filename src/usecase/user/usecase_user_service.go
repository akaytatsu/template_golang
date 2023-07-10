package usecase_user

import "app/entity"

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
