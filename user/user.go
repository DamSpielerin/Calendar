package user

import "github.com/dgrijalva/jwt-go"

type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Timezone string `json:"timezone"`
}

type JwtToken struct {
	Token string `json:"token"`
}

type Exception struct {
	Message string `json:"message"`
}
type AccessDetails struct {
	AccessUuid string
	UserId     int
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
// We add jwt.StandardClaims as an embedded type, to provide fields like expiry time
type Claims struct {
	Username string `json:"username"`
	Timezone string `json:"timezone"`
	jwt.StandardClaims
}
