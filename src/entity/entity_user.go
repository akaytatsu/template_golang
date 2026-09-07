package entity

import (
	"app/config"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidToken indica um token cuja assinatura ou estrutura não confere.
	ErrInvalidToken = errors.New("invalid token")
	// ErrInvalidClaims indica um token válido cujas claims não são SignedDetails.
	ErrInvalidClaims = errors.New("invalid token claims")
)

// signingMethod é o único algoritmo aceito na validação de tokens. Restringir
// a lista evita ataques de algorithm confusion.
var signingMethod = jwt.SigningMethodHS256

type SignedDetails struct {
	ID    int
	Name  string
	Email string
	jwt.RegisteredClaims
}

type EntityUserFilters struct {
	IDs    []uint `json:"ids"`
	Search string `json:"search"`
	Active string `json:"active"`

	// Page é 0-indexado. PageSize <= 0 desliga a paginação.
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type EntityUser struct {
	ID        int
	Name      string    `json:"name"       validate:"required,min=3,max=120"`
	Email     string    `json:"email"      validate:"required,email"`
	Password  string    `json:"password"   validate:"required,min=4,max=120"`
	IsAdmin   bool      `json:"is_admin" gorm:"default:false"`
	Active    bool      `json:"active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MarshalJSON serializa o usuário omitindo Password: o hash bcrypt não deve
// sair em nenhuma resposta da API (antes /api/user/me e /api/user/list o
// devolviam). O unmarshal segue usando as tags normais, então os payloads de
// criação e atualização continuam aceitando `password`.
func (u EntityUser) MarshalJSON() ([]byte, error) {
	type publicUser struct {
		ID        int       `json:"ID"`
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		IsAdmin   bool      `json:"is_admin"`
		Active    bool      `json:"active"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	return json.Marshal(publicUser{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		IsAdmin:   u.IsAdmin,
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	})
}

// NewUser valida a senha em texto puro e devolve um usuário com o hash já
// aplicado. O EntityUser retornado NÃO deve passar por GetValidated, que
// hashearia a senha uma segunda vez.
func NewUser(userParam EntityUser) (*EntityUser, error) {
	now := time.Now()

	if err := ValidateRawPassword(userParam.Password); err != nil {
		return nil, err
	}

	password, err := GeneratePassword(userParam.Password)
	if err != nil {
		return nil, err
	}

	u := &EntityUser{
		Name:      userParam.Name,
		Email:     userParam.Email,
		Password:  password,
		IsAdmin:   userParam.IsAdmin,
		Active:    userParam.Active,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return u, nil
}

func (u *EntityUser) ValidatePassword(p string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(p))
	if err != nil {
		return err
	}

	return nil
}

func (u *EntityUser) Validate() error {
	return validate.Struct(u)
}

// UpdatePassword valida a senha em texto puro e substitui u.Password pelo hash.
// Como a senha já sai hasheada, NÃO chame GetValidated depois desta função — o
// hash seria hasheado de novo e a senha nova nunca autenticaria.
func (u *EntityUser) UpdatePassword(newPassword string) error {
	if err := ValidateRawPassword(newPassword); err != nil {
		return err
	}

	hash, err := GeneratePassword(newPassword)
	if err != nil {
		return err
	}

	u.Password = hash

	return nil
}

// GetValidated valida o usuário e hasheia u.Password em texto puro.
//
// ATENÇÃO: não é idempotente. Chamar duas vezes gera bcrypt(bcrypt(senha)) e
// quebra a autenticação. Use apenas no caminho de criação, sobre um EntityUser
// que ainda tem a senha em texto puro.
func (u *EntityUser) GetValidated() error {
	err := u.Validate()
	if err != nil {
		return err
	}

	pwd, err := GeneratePassword(u.Password)
	if err != nil {
		return err
	}
	u.Password = pwd

	return nil
}

func (u *EntityUser) JWTTokenGenerator() (signedToken string, signedRefreshToken string, err error) {
	claims := SignedDetails{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	refreshClaims := SignedDetails{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7 * 365)),
		},
	}

	token, err := jwt.NewWithClaims(signingMethod, claims).SignedString([]byte(config.EnvironmentVariables.JWT_SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(signingMethod, refreshClaims).SignedString([]byte(config.EnvironmentVariables.JWT_SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

// ValidateToken devolve as claims de um token assinado. A expiração é validada
// pelo próprio jwt/v5, que retorna jwt.ErrTokenExpired.
func (u *EntityUser) ValidateToken(signedToken string) (*SignedDetails, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (any, error) {
			return []byte(config.EnvironmentVariables.JWT_SECRET_KEY), nil
		},
		jwt.WithValidMethods([]string{signingMethod.Alg()}),
	)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

func GeneratePassword(raw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
