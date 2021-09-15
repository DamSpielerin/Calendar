package user

import (
	"calendar/event"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID       uuid.UUID     `json:"id,omitempty" gorm:"primaryKey;"`
	Login    string        `json:"login"`
	Email    string        `json:"email,omitempty"`
	Password string        `json:"password,omitempty" gorm:"column:password_hash"`
	Timezone string        `json:"timezone,omitempty"`
	Events   []event.Event `gorm:"foreignKey:UserId"`
}

// Create the JWT key used to create the signature
var JwtKey = []byte("my_secret_key")

// Create a struct to read the username and password from the request body
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
	Timezone string `json:"timezone,omitempty"`
}

// Create a struct that will be encoded to a JWT.
type Claims struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Timezone string    `json:"timezone"`
	jwt.StandardClaims
}
