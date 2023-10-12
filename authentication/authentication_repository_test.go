package authentication

import (
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticationRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}

	authRepo := &AuthRepository{
		DB: db,
	}

	defer db.Close()

	t.Run("Create JWT token", func(t *testing.T) {
		os.Setenv("JWT_AUTH_SECRET", "My_secret")
		defer os.Unsetenv("JWT_AUTH_SECRET")
		t.Run("verifies that the JWT token is generated", func(t *testing.T) {
			username := "testuser"
			userID := "123456"

			token, err := CreateJWTToken(username, userID)

			if err != nil {
				t.Errorf("Error creating JWT token: %v", err)
			}

			if token == "" {
				t.Error("JWT token is empty")
			}
		})

		t.Run("verifies that the JWT token is not generated when username is empty", func(t *testing.T) {
			username := ""
			userID := "123456"

			token, err := CreateJWTToken(username, userID)

			if err == nil {
				t.Errorf("invalid username or user_id")
			}

			if token != "" {
				t.Error("JWT token is not empty")
			}
		})

		t.Run("verifies that the JWT token is not generated when user_id is empty", func(t *testing.T) {
			username := "testuser"
			userID := ""

			token, err := CreateJWTToken(username, userID)

			if err == nil {
				t.Errorf("invalid username or user_id")
			}

			if token != "" {
				t.Error("JWT token is not empty")
			}
		})
	})

	t.Run("Test Get user from token", func(t *testing.T) {
		os.Setenv("JWT_AUTH_SECRET", "My_secret")
		defer os.Unsetenv("JWT_AUTH_SECRET")
		t.Run("verifies that the user_id is retrieved from a jwt token", func(t *testing.T) {
			userId := uuid.New().String()

			token := &jwt.Token{
				Claims: jwt.MapClaims{
					"user_id": userId,
				},
			}

			jwtUserId, ok := GetUserFromToken(token)
			if !ok {
				t.Error("Expected ok to be true, got false")
			}
			if jwtUserId != userId {
				t.Errorf("Expected userId to be '%s', got '%s'", userId, jwtUserId)
			}
		})

		t.Run("verifies that the error is thrown when failed to retreive user_id", func(t *testing.T) {
			token := &jwt.Token{
				Claims: jwt.MapClaims{},
			}

			_, ok := GetUserFromToken(token)

			if ok {
				t.Error("Expected ok to be false, got true")
			}
		})

		t.Run(("verifies that the user cannot be retreived when the token is nil"), func(t *testing.T) {
			var token *jwt.Token

			_, ok := GetUserFromToken(token)

			if ok {
				t.Error("Expected ok to be false, got true")
			}
		})
	})

	t.Run("check email exisists", func(t *testing.T) {
		email := "guest@gmail.com"
		expectedQuery := `SELECT EXISTS \(SELECT 1 FROM users WHERE email = 'guest@gmail.com'\) AS email_exists`

		t.Run("Verifies that the email already exists", func(t *testing.T) {
			rows := mock.NewRows([]string{"email_exists"}).AddRow(true)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			got, err := authRepo.checkEmailExists(email)

			if err != nil {
				t.Fatalf("Error executing email already exists: %v", err)
			}

			assert.True(t, got, "Expected the email already exists")
			assert.NoError(t, mock.ExpectationsWereMet(), "Expectations were not met")
		})

		t.Run("Verifies that the email does not exists", func(t *testing.T) {
			rows := mock.NewRows([]string{"email_exists"}).AddRow(false)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			got, err := authRepo.checkEmailExists(email)

			if err != nil {
				t.Fatalf("Error executing email already exists: %v", err)
			}

			if got {
				t.Errorf("Email already exists")
			}

			assert.False(t, got, "Expected the email not exists")
			assert.NoError(t, mock.ExpectationsWereMet(), "Expectations were not met")
		})

		t.Run("Verifies that the error is thrown when cannot find if the email exists", func(t *testing.T) {
			err := errors.New("internal server error")
			mock.ExpectQuery(expectedQuery).WillReturnError(err)

			_, err = authRepo.checkEmailExists(email)

			if err == nil {
				t.Fatalf("Error must be thrown")
			}

			assert.NoError(t, mock.ExpectationsWereMet(), "Expectations were not met")
		})
	})

	t.Run("Verify the register service", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"

		t.Run("verifies the user register service us called successfully", func(t *testing.T) {
			mock.ExpectExec("INSERT INTO users").
				WithArgs(sqlmock.AnyArg(), email, sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err = authRepo.Register(email, password)

			assert.NoError(t, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		t.Run("verifies the error is thrown when user can not be registered", func(t *testing.T) {
			mock.ExpectExec("INSERT INTO users").
				WithArgs(sqlmock.AnyArg(), email, sqlmock.AnyArg()).
				WillReturnError(sql.ErrNoRows)

			err := authRepo.Register(email, password)
			assert.Error(t, err)

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})

	t.Run("Login", func(t *testing.T) {
		email := "guest@gmail.com"
		expectedQuery := `SELECT password AS hashPassword, id AS userId FROM users WHERE email = 'guest@gmail.com'`
		t.Run("Verifies that login is successful", func(t *testing.T) {
			rows := mock.NewRows([]string{"hashPassword", "userId"}).AddRow("hashPassword", "user-id")
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			hashPassword, userId, err := authRepo.Login(email)

			if err != nil {
				t.Fatalf("Error must not be thrown: %v", err)
			}

			if hashPassword != "hashPassword" || userId != "user-id" {
				t.Errorf("error is not expected")
			}
		})

		t.Run("Verifies that login is successful", func(t *testing.T) {
			err := errors.New("internal server error")
			mock.ExpectQuery(expectedQuery).WillReturnError(err)

			hashPassword, userId, err := authRepo.Login(email)

			if err == nil {
				t.Fatalf("Error must be thrown: %v", err)
			}

			if hashPassword == "hashPassword" || userId == "user-id" {
				t.Errorf("error is not expected")
			}
		})
	})
}
