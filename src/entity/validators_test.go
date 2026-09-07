package entity_test

import (
	"app/entity"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regressão: GetStructError fazia uma type assertion sem `ok`, então qualquer
// erro que não viesse do validator causava panic.
func TestGetStructError_NonValidationError(t *testing.T) {
	assert.NotPanics(t, func() {
		assert.Empty(t, entity.GetStructError(errors.New("erro qualquer")))
	})
}

func TestGetStructError_NilError(t *testing.T) {
	assert.NotPanics(t, func() {
		assert.Empty(t, entity.GetStructError(nil))
	})
}

// Regressão: err.Value().(string) estourava em campo não-string. IsAdmin é
// bool, então um erro de validação nele derrubava o processo.
func TestGetStructError_NonStringField(t *testing.T) {
	user := entity.EntityUser{
		Name:    "", // required: falha
		Email:   "nao-e-email",
		IsAdmin: true,
	}

	err := user.Validate()
	require.Error(t, err)

	var fields []entity.IError

	assert.NotPanics(t, func() {
		fields = entity.GetStructError(err)
	})

	require.NotEmpty(t, fields)

	byField := make(map[string]entity.IError, len(fields))
	for _, f := range fields {
		byField[f.Field] = f
	}

	require.Contains(t, byField, "Name")
	assert.Equal(t, "required", byField["Name"].Tag)
}

func TestValidateRawPassword(t *testing.T) {
	assert.NoError(t, entity.ValidateRawPassword("senha-ok"))
	assert.Error(t, entity.ValidateRawPassword(""))
	assert.Error(t, entity.ValidateRawPassword("abc"))
	assert.Error(t, entity.ValidateRawPassword(string(make([]byte, 121))))
}

// Regressão de segurança: a API devolvia o hash bcrypt no corpo da resposta
// (GET /api/user/me e /api/user/list). O unmarshal precisa continuar aceitando
// `password`, senão a criação de usuário para de funcionar.
func TestEntityUser_MarshalJSONOmitsPassword(t *testing.T) {
	user := entity.EntityUser{
		ID:       1,
		Name:     "Name",
		Email:    "user@example.com",
		Password: "$2a$10$hash-que-nao-pode-vazar",
		IsAdmin:  true,
	}

	raw, err := json.Marshal(user)
	require.NoError(t, err)

	assert.NotContains(t, string(raw), "hash-que-nao-pode-vazar")
	assert.NotContains(t, string(raw), "password")
	// Os demais campos seguem presentes.
	assert.Contains(t, string(raw), "user@example.com")
	assert.Contains(t, string(raw), `"is_admin":true`)
}

func TestEntityUser_MarshalJSONInsideSliceAndPointer(t *testing.T) {
	users := []entity.EntityUser{{ID: 1, Password: "segredo-a"}}
	raw, err := json.Marshal(users)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "segredo-a")

	raw, err = json.Marshal(&entity.EntityUser{ID: 2, Password: "segredo-b"})
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "segredo-b")
}

func TestEntityUser_UnmarshalStillReadsPassword(t *testing.T) {
	var user entity.EntityUser

	err := json.Unmarshal([]byte(`{"name":"N","email":"e@e.com","password":"senhaCrua1"}`), &user)
	require.NoError(t, err)

	assert.Equal(t, "senhaCrua1", user.Password, "o bind de criação depende disso")
}
