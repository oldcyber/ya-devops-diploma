package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	// "git.oldcyber.xyz/poip-it/bdd/internal/app/apiserver/session"
	"github.com/golang-jwt/jwt"
	log "github.com/sirupsen/logrus"

	"net/http"
	"strings"
)

const (
	apiKey      = "fjvGxBP621wBemoc4ukb8IhV9ku9oAPi6xRYyCrbc7df1Z3MIPTM2LdG20PZVAxE" //nolint:gosec
	tokenExpire = "24h"
)

// CreateToken creates a new token
func CreateToken(userID int) (string, error) {
	var expTime, err = time.ParseDuration(tokenExpire)
	if err != nil {
		log.Println(err)
	}
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(expTime).Unix() // Token expires after expTime hour
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken, _ := token.SignedString([]byte(apiKey))
	return jwtToken, nil
}

// TokenValid validates token
func TokenValid(r *http.Request) (string, error) {
	var ruid string
	tokenString := extractToken(r)
	if tokenString == "" {
		return "", errors.New("missing token")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(apiKey), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		uid, err := strconv.ParseUint(fmt.Sprintf("%.0f", claims["user_id"]), 10, 32)
		if err != nil {
			return "", err
		}
		ruid = strconv.FormatUint(uid, 10)
	}
	return ruid, nil
}

// extractToken extracts token from Authorization header
func extractToken(r *http.Request) string {
	keys := r.URL.Query()
	token := keys.Get("token")
	if token != "" {
		return token
	}
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}
