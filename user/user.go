package user

import "github.com/dgrijalva/jwt-go"

type User struct {
	ID       int    `json:"id,omitempty"`
	Login    string `json:"login"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	Timezone string `json:"timezone,omitempty"`
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
	Username string `json:"username"`
	Timezone string `json:"timezone"`
	jwt.StandardClaims
}
