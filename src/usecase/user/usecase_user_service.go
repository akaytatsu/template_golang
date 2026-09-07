package usecase_user

import (
	"app/entity"
	"errors"

	"gorm.io/gorm"
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

// Update persiste o usuário. O repositório usa Save (grava todas as colunas),
// então a senha precisa de tratamento explícito: vinda em texto puro no payload
// ela seria gravada sem hash. Payload sem senha preserva o hash atual.
func (u *UseCaseUser) Update(user *entity.EntityUser) error {
	current, err := u.repo.GetByID(user.ID)
	if err != nil {
		return err
	}

	if current == nil {
		return gorm.ErrRecordNotFound
	}

	if user.Password == "" {
		user.Password = current.Password
	} else if err := user.UpdatePassword(user.Password); err != nil {
		return err
	}

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

	// UpdatePassword já valida a senha em texto puro e aplica o bcrypt. Chamar
	// GetValidated aqui hashearia o hash e a senha nova nunca autenticaria.
	if err := user.UpdatePassword(newPassword); err != nil {
		return err
	}

	return u.repo.UpdateUser(user)
}

func (u *UseCaseUser) GetUsersFromIDs(ids []int) (users []entity.EntityUser, err error) {
	return u.repo.GetUsersFromIDs(ids)
}

func (u *UseCaseUser) GetUsers(filters entity.EntityUserFilters) (users []entity.EntityUser, total int64, err error) {
	return u.repo.GetUsers(filters)
}

func (u *UseCaseUser) GetUser(id int) (user *entity.EntityUser, err error) {
	return u.repo.GetUser(id)
}

func JWTTokenGenerator(u entity.EntityUser) (signedToken string, signedRefreshToken string, err error) {
	return u.JWTTokenGenerator()
}
