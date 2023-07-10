package entity

import (
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

const SECRET_KEY = "9an0afx$thw)k9#y*_d9-ch^r&a6ndi#x#dwu^52zbqw=hso(9"

type SignedDetails struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	jwt.StandardClaims
}

type EntityUser struct {
	ID        int
	FirstName string    `json:"first_name" validate:"required,min=2,max=50"`
	LastName  string    `json:"last_name"  validate:"required,min=2,max=50"`
	Email     string    `json:"email"      validate:"required,email"`
	Password  string    `json:"password"   validate:"required,min=8,max=120"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUser(userParam EntityUser) (*EntityUser, error) {

	now := time.Now()

	password, err := GeneratePassword(userParam.Password)

	if err != nil {
		return nil, err
	}

	u := &EntityUser{
		FirstName: userParam.FirstName,
		LastName:  userParam.LastName,
		Email:     userParam.Email,
		Password:  password,
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
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}

	refreshClaims := SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7 * 365).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

func (u *EntityUser) ValidateToken(signedToken string) (claims *SignedDetails, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {

		return nil, err
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {

		return nil, err
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {

		return nil, err
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
