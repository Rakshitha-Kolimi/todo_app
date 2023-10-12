package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func createJwtToken(username, id string) (*jwt.Token, string) {
	claims := jwt.MapClaims{
		"name":    username,
		"user_id": id,
		"exp":     jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}

	jwtSecretKey := os.Getenv("JWT_AUTH_SECRET")

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	rawToken, err := token.SignedString([]byte(jwtSecretKey))
	if err != nil {
		return nil, ""
	}

	return token, rawToken
}

func TestController(t *testing.T) {
	e := echo.New()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}

	username := "example@gmail.com"
	user_id := "ed6caeda-1fa9-442e-a41d-dd2b135cea67"

	token, rawToken := createJwtToken(username, user_id)

	req := httptest.NewRequest(http.MethodGet, "/item/main", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer"+" "+rawToken)

	rec := httptest.NewRecorder()

	ctx := e.NewContext(req, rec)
	ctx.Set("user", token)

	controller(db, e)

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	assert.Equal(t, "Welcome to main page", rec.Body.String())
}
