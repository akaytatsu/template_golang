package entity_test

import (
	"app/config"
	"app/entity"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityUser_NewUserHashesPassword(t *testing.T) {
	u := entity.EntityUser{
		Name:     "Name",
		Email:    "email@email.com",
		Password: "passwordTest",
	}

	user, err := entity.NewUser(u)
	require.NoError(t, err)

	// A senha tem de sair hasheada, não em texto puro.
	assert.NotEqual(t, u.Password, user.Password)
	assert.NoError(t, user.ValidatePassword(u.Password))
}

func TestEntityUser_NewUserRejectsEmptyPassword(t *testing.T) {
	_, err := entity.NewUser(entity.EntityUser{
		Name:  "Name",
		Email: "email@email.com",
	})

	assert.Error(t, err, "senha vazia deve ser rejeitada, não hasheada")
}

func TestEntityUser_ValidatedSuccess(t *testing.T) {
	arg := entity.EntityUser{
		Name:     "Name",
		Email:    "email@email.com",
		Password: "Password",
	}

	user, err := entity.NewUser(arg)
	require.NoError(t, err)

	assert.NoError(t, user.Validate())
}

func TestEntityUser_ValidatedFail(t *testing.T) {
	// Campos obrigatórios em branco reprovam na validação da struct.
	user := entity.EntityUser{
		Name:     "",
		Email:    "",
		Password: "senha-valida",
	}

	assert.Error(t, user.GetValidated())
}

// Regressão do bug de double-bcrypt: UpdatePassword hasheia uma única vez, e a
// senha nova precisa autenticar depois.
func TestEntityUser_UpdatePasswordHashesOnce(t *testing.T) {
	user := entity.EntityUser{
		Name:     "Name",
		Email:    "email@email.com",
		Password: "senhaAntiga",
	}

	require.NoError(t, user.UpdatePassword("senhaNova123"))

	assert.NoError(t, user.ValidatePassword("senhaNova123"))
	assert.Error(t, user.ValidatePassword("senhaAntiga"))
}

func TestEntityUser_UpdatePasswordRejectsInvalid(t *testing.T) {
	user := entity.EntityUser{Password: "senhaAntiga"}

	for name, newPassword := range map[string]string{
		"vazia": "",
		"curta": "abc",
		"longa": string(make([]byte, 121)),
	} {
		t.Run(name, func(t *testing.T) {
			assert.Error(t, user.UpdatePassword(newPassword))
		})
	}
}

func TestEntityUser_ValidateToken(t *testing.T) {
	config.EnvironmentVariables.JWT_SECRET_KEY = "chave-de-teste"

	user := entity.EntityUser{ID: 42, Name: "Name", Email: "email@email.com"}

	token, refreshToken, err := user.JWTTokenGenerator()
	require.NoError(t, err)
	require.NotEmpty(t, refreshToken)

	claims, err := user.ValidateToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, 42, claims.ID)
	assert.Equal(t, "email@email.com", claims.Email)
}

// Regressão: antes estes caminhos devolviam (nil, nil) e o chamador tratava
// como sucesso, com claims nil.
func TestEntityUser_ValidateTokenNeverReturnsNilNil(t *testing.T) {
	config.EnvironmentVariables.JWT_SECRET_KEY = "chave-de-teste"

	var user entity.EntityUser

	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, entity.SignedDetails{
		ID: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	})
	expiredToken, err := expired.SignedString([]byte("chave-de-teste"))
	require.NoError(t, err)

	otherKey := jwt.NewWithClaims(jwt.SigningMethodHS256, entity.SignedDetails{ID: 1})
	wrongKeyToken, err := otherKey.SignedString([]byte("outra-chave"))
	require.NoError(t, err)

	for name, token := range map[string]string{
		"vazio":        "",
		"malformado":   "not-a-jwt",
		"com o header": "Bearer abc.def.ghi",
		"expirado":     expiredToken,
		"chave errada": wrongKeyToken,
	} {
		t.Run(name, func(t *testing.T) {
			claims, err := user.ValidateToken(token)

			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestEntityUser_ValidateTokenRejectsUnexpectedAlg(t *testing.T) {
	config.EnvironmentVariables.JWT_SECRET_KEY = "chave-de-teste"

	// alg=none não deve ser aceito (algorithm confusion).
	token := jwt.NewWithClaims(jwt.SigningMethodNone, entity.SignedDetails{ID: 1})
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	var user entity.EntityUser

	claims, err := user.ValidateToken(signed)

	assert.Error(t, err)
	assert.Nil(t, claims)
}
