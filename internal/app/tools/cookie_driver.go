package tools

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	JWTSecretKey = "5289121a-f775-477f-a18d-8e937e9d93be"
	JWTCookieKey = "token"
)

// UserClaims is a custom JWT claims structure
type UserClaims struct {
	UserID string `json:"userID"`
	jwt.RegisteredClaims
}

func keyFunc(token *jwt.Token) (interface{}, error) {
	_, ok := token.Method.(*jwt.SigningMethodHMAC)
	if !ok {
		err := fmt.Errorf("unexpected singing method %v", token.Header["alg"])
		return nil, err
	}
	return []byte(JWTSecretKey), nil
}

func GetTokenAndUserID(cookie *http.Cookie) (*jwt.Token, int, error) {
	claims := &UserClaims{}
	token, err := jwt.ParseWithClaims(cookie.Value, claims, keyFunc)
	if err != nil || !token.Valid {
		return token, 0, err
	}

	userID, err := strconv.Atoi(claims.UserID)
	if err != nil {
		return token, 0, err
	}

	return token, userID, nil
}

func SetUserIDCookie(writer http.ResponseWriter, request *http.Request, userID string) {
	claims := &UserClaims{
		userID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "myServer",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(JWTSecretKey))
	if err != nil {
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	newCookie := &http.Cookie{
		Name:    JWTCookieKey,
		Value:   signedToken,
		Expires: time.Now().Add(24 * time.Hour),
	}

	request.AddCookie(newCookie)

	http.SetCookie(writer, newCookie)
}
