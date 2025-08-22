package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TokenType string

const (
	AccessTokenType TokenType = "chirpy-issuer"
)

var ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userId uuid.UUID, tokenSecret string) (string, error) {

	expiresIn := time.Duration(time.Minute * 60)
	signinKey := []byte(tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(AccessTokenType),
		Subject:   userId.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})

	tokenStr, err := token.SignedString(signinKey)
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)
	hexStr := hex.EncodeToString(key)
	fmt.Println("hex.EncodeToString output: ", hexStr)
	return hexStr
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claimsStruct, func(t *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	userIdStr, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}
	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}

	if issuer != string(AccessTokenType) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}

	expiresAt, err := token.Claims.GetExpirationTime()
	if err != nil {
		return uuid.Nil, err
	}

	if expiresAt.Compare(time.Now()) <= 0 {
		return uuid.Nil, errors.New("token is expired")
	}

	return userId, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if len(authHeader) == 0 {
		return "", errors.New("authorization headers not provided")
	}

	authHeaderList := strings.Split(authHeader, " ")
	if len(authHeaderList) < 2 || authHeaderList[0] != "Bearer" {
		return "", errors.New("malformed authorization header")
	}

	return authHeaderList[1], nil
}

func GetUserIdFromReq(r *http.Request, secret string) (uuid.UUID, error) {
	token, err := GetBearerToken(r.Header)
	if err != nil {
		return uuid.Nil, errors.New("no bearer authentication header provided")
	}
	log.Println("GetingUserIdFromReq token: ", token)

	userId, err := ValidateJWT(token, secret)
	if err != nil {
		log.Println(err)
		return uuid.Nil, errors.New("invalid jwt token provided")
	}

	return userId, nil
}
