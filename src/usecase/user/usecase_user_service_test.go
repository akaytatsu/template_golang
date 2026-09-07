package usecase_user_test

import (
	"app/entity"
	"app/mocks"
	"errors"
	"testing"

	usecase_user "app/usecase/user"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUsecaseUser_LoginUser(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	password, _ := entity.GeneratePassword("password33")

	mockUserRepo := mocks.NewMockIRepositoryUser(ctrl)
	mockUserRepo.EXPECT().GetByMail(gomock.Any()).Return(&entity.EntityUser{
		Email:    "mailer@mailer.com",
		Password: password,
	}, nil)

	_, err := usecase_user.NewService(mockUserRepo).LoginUser("mailer@mailer.com", "password33")

	assert.Nil(t, err)
}

func TestUsecaseUser_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockIRepositoryUser(ctrl)
	mockUserRepo.EXPECT().CreateUser(gomock.Any()).Return(nil)

	Convey("User can't be created", t, func() {
		err := usecase_user.NewService(mockUserRepo).Create(&entity.EntityUser{})

		So(err, ShouldNotBeNil)
	})

	Convey("User can be created", t, func() {
		err := usecase_user.NewService(mockUserRepo).Create(&entity.EntityUser{
			Email:    "mailer@mailer.com",
			Name:     "Name",
			Password: "password33",
		})

		So(err, ShouldBeNil)
	})
}

// Create hasheia a senha exatamente uma vez, então ela precisa autenticar.
func TestUsecaseUser_CreateHashesPasswordOnce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var persisted *entity.EntityUser

	mockUserRepo := mocks.NewMockIRepositoryUser(ctrl)
	mockUserRepo.EXPECT().CreateUser(gomock.Any()).
		DoAndReturn(func(u *entity.EntityUser) error {
			persisted = u
			return nil
		})

	err := usecase_user.NewService(mockUserRepo).Create(&entity.EntityUser{
		Name:     "Name",
		Email:    "mailer@mailer.com",
		Password: "password33",
	})
	require.NoError(t, err)
	require.NotNil(t, persisted)

	assert.NotEqual(t, "password33", persisted.Password)
	assert.NoError(t, persisted.ValidatePassword("password33"))
}

// Regressão do bug mais grave: UpdatePassword aplicava bcrypt duas vezes
// (UpdatePassword + GetValidated), então a senha nova nunca autenticava.
func TestUsecaseUser_UpdatePasswordAuthenticatesWithNewPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldHash, err := entity.GeneratePassword("senhaAntiga1")
	require.NoError(t, err)

	stored := &entity.EntityUser{
		ID:       1,
		Name:     "Name",
		Email:    "mailer@mailer.com",
		Password: oldHash,
	}

	var persisted *entity.EntityUser

	mockUserRepo := mocks.NewMockIRepositoryUser(ctrl)
	mockUserRepo.EXPECT().GetByID(1).Return(stored, nil)
	mockUserRepo.EXPECT().UpdateUser(gomock.Any()).
		DoAndReturn(func(u *entity.EntityUser) error {
			persisted = u
			return nil
		})

	err = usecase_user.NewService(mockUserRepo).
		UpdatePassword(1, "senhaAntiga1", "senhaNova123", "senhaNova123")
	require.NoError(t, err)
	require.NotNil(t, persisted)

	assert.NoError(t, persisted.ValidatePassword("senhaNova123"),
		"a senha nova tem de autenticar — bcrypt aplicado uma única vez")
	assert.Error(t, persisted.ValidatePassword("senhaAntiga1"))
}

func TestUsecaseUser_UpdatePasswordValidations(t *testing.T) {
	oldHash, err := entity.GeneratePassword("senhaAntiga1")
	require.NoError(t, err)

	cases := map[string]struct {
		old, new, confirm string
	}{
		"senha antiga errada":   {"errada", "senhaNova123", "senhaNova123"},
		"confirmacao diferente": {"senhaAntiga1", "senhaNova123", "outraCoisa12"},
		"senha nova curta":      {"senhaAntiga1", "abc", "abc"},
		"senha nova vazia":      {"senhaAntiga1", "", ""},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mocks.NewMockIRepositoryUser(ctrl)
			mockUserRepo.EXPECT().GetByID(1).Return(&entity.EntityUser{
				ID:       1,
				Name:     "Name",
				Email:    "mailer@mailer.com",
				Password: oldHash,
			}, nil)
			// Nenhum destes casos deve chegar a persistir.

			err := usecase_user.NewService(mockUserRepo).
				UpdatePassword(1, tc.old, tc.new, tc.confirm)

			assert.Error(t, err)
		})
	}
}

// Update usa Save (grava todas as colunas). Payload sem senha deve preservar o
// hash atual em vez de zerá-lo.
func TestUsecaseUser_UpdatePreservesPasswordWhenAbsent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	existingHash, err := entity.GeneratePassword("senhaAtual12")
	require.NoError(t, err)

	var persisted *entity.EntityUser

	mockUserRepo := mocks.NewMockIRepositoryUser(ctrl)
	mockUserRepo.EXPECT().GetByID(1).Return(&entity.EntityUser{ID: 1, Password: existingHash}, nil)
	mockUserRepo.EXPECT().UpdateUser(gomock.Any()).
		DoAndReturn(func(u *entity.EntityUser) error {
			persisted = u
			return nil
		})

	err = usecase_user.NewService(mockUserRepo).Update(&entity.EntityUser{
		ID:    1,
		Name:  "Novo Nome",
		Email: "novo@example.com",
	})
	require.NoError(t, err)
	require.NotNil(t, persisted)

	assert.Equal(t, existingHash, persisted.Password)
	assert.NoError(t, persisted.ValidatePassword("senhaAtual12"))
}

// Payload com senha: precisa ser hasheada, nunca gravada em texto puro.
func TestUsecaseUser_UpdateHashesProvidedPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var persisted *entity.EntityUser

	mockUserRepo := mocks.NewMockIRepositoryUser(ctrl)
	mockUserRepo.EXPECT().GetByID(1).Return(&entity.EntityUser{ID: 1, Password: "hash-antigo"}, nil)
	mockUserRepo.EXPECT().UpdateUser(gomock.Any()).
		DoAndReturn(func(u *entity.EntityUser) error {
			persisted = u
			return nil
		})

	err := usecase_user.NewService(mockUserRepo).Update(&entity.EntityUser{
		ID:       1,
		Name:     "Nome",
		Email:    "novo@example.com",
		Password: "senhaNova123",
	})
	require.NoError(t, err)
	require.NotNil(t, persisted)

	assert.NotEqual(t, "senhaNova123", persisted.Password, "senha não pode ir em texto puro para o banco")
	assert.NoError(t, persisted.ValidatePassword("senhaNova123"))
}

func TestUsecaseUser_GetUserByTokenPropagatesRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Token inválido: o repositório não deve nem ser consultado.
	mockUserRepo := mocks.NewMockIRepositoryUser(ctrl)

	user, err := usecase_user.NewService(mockUserRepo).GetUserByToken("token-invalido")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUsecaseUser_GetUsersPropagatesTotal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockIRepositoryUser(ctrl)
	mockUserRepo.EXPECT().GetUsers(gomock.Any()).
		Return([]entity.EntityUser{{ID: 1}}, int64(42), nil)

	users, total, err := usecase_user.NewService(mockUserRepo).
		GetUsers(entity.EntityUserFilters{Page: 0, PageSize: 10})

	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, int64(42), total)
}

func TestUsecaseUser_LoginUserPropagatesRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockIRepositoryUser(ctrl)
	mockUserRepo.EXPECT().GetByMail("nao@existe.com").Return(nil, errors.New("record not found"))

	user, err := usecase_user.NewService(mockUserRepo).LoginUser("nao@existe.com", "qualquer")

	assert.Error(t, err)
	assert.Nil(t, user)
}
