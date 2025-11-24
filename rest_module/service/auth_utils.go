package service

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	. "rest_module/model"
)

var JWT_SECRET = "53A73E5F1C4E0A2D3B5F2D784E6A1B423D6F247D1F6E5C3A596D635A75327855"

func CheckTokenAndGetId(tokenString string) (string, error) {
	log.Println("Проверка JWT токена")
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (any, error) {
			return []byte(JWT_SECRET), nil
		})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("Не валидный токен аутентификации %s", err.Error())
	}

	return claims.ID, nil
}

func GenerateJWTToken(id string) (string, error) {
	log.Println("Создание JWT токена")
	claims := jwt.RegisteredClaims{
		ID:        id,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWT_SECRET))
}

func CheckPasswordForUser(user *User, password string) error {
	log.Println("Проверка пароля")
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	// Проверяем наличие ошибок
	if err != nil {
		return fmt.Errorf("Не правильный пароль %s", err.Error())
	}

	return nil
}
