package user

import (
	"calendar/event"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID           uuid.UUID     `json:"id,omitempty" gorm:"primaryKey;"`
	Login        string        `json:"login"`
	Email        string        `json:"email,omitempty"`
	PasswordHash string        `json:"password,omitempty" gorm:"column:password_hash"`
	Timezone     string        `json:"timezone,omitempty"`
	Events       []event.Event `gorm:"foreignKey:UserId"`
}

// Create the JWT key used to create the signature
var JwtKey = []byte("my_secret_key")

// Create a struct to read the username and password from the request body
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
	Timezone string `json:"timezone,omitempty"`
	Email    string `json:"email,omitempty""`
}

// Create a struct that will be encoded to a JWT.
type Claims struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Timezone string    `json:"timezone"`
	jwt.StandardClaims
}

//PasswordHash create
func (cred *Credentials) PasswordHash() (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(cred.Password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (cred *Credentials) PasswordVerify(hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(cred.Password))
	return err == nil
}
