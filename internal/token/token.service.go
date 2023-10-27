package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var _ TokenService = &tokenServiceImpl{}

type tokenServiceImpl struct {
	secret []byte
	ttl    int
}

type TokenService interface {
	CreateToken(uid string) string
	Validate(tokenString string) (string, error)
	TTL() int
}

func NewService(secret []byte, ttl int) TokenService {
	return &tokenServiceImpl {
		secret,
		ttl,
	}
}

var (
	issuer = "paotooong"
)

func (s *tokenServiceImpl) CreateToken(uid string) string {
	ss := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": issuer,
		"sub": uid,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Second * time.Duration(s.ttl)).Unix(),
	})
	token, _ := ss.SignedString(s.secret)
	return token
}

func (s *tokenServiceImpl) Validate(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return s.secret, nil
	}, jwt.WithIssuer(issuer))

	if err != nil {
		return "", errors.New("error while validating token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if uid, ok := claims["sub"].(string); ok {
			return uid, nil
		}
	}
	return "", errors.New("invalid token")
}


func (s *tokenServiceImpl) TTL() int {
	return s.ttl
}
