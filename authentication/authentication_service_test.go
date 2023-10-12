package authentication

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"todo-project/constants"
)

func TestAuthenticationService(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}

	authRepo := &AuthRepository{
		DB: db,
	}

	as := AuthService{
		R: *authRepo,
	}

	defer db.Close()
	t.Run("User Register", func(t *testing.T) {
		t.Run("verifies that the user is registered successfully", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM users WHERE email = 'johndoe@gmail.com'\) AS email_exists`

			rows := mock.NewRows([]string{"email_exists"}).AddRow(false)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			mock.ExpectExec("INSERT INTO users").
				WithArgs(sqlmock.AnyArg(), "johndoe@gmail.com", sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))

			e := echo.New()

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"email": "johndoe@gmail.com", "password": "JohnD0@2123"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err = as.UserRegister(c)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Equal(t, constants.USER_REGISTERED_SUCCESSFUL, rec.Body.String())
		})

		t.Run("verifies that the email already exsists error is thrown", func(t *testing.T) {
			e := echo.New()

			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM users WHERE email = 'johndoe@gmail.com'\) AS email_exists`

			rows := mock.NewRows([]string{"email_exists"}).AddRow(true)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"email": "johndoe@gmail.com", "password": "JohnD0@2123"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err = as.UserRegister(c)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, rec.Code)
			assert.Equal(t, constants.EMAIL_ADDRESS_ALREADY_EXISTS, rec.Body.String())
		})

		t.Run("verifies that the internal server error is thrown", func(t *testing.T) {
			e := echo.New()

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"email": "johndoe@gmail.com", "password": "JohnD0@2123"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err = as.UserRegister(c)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		})

		t.Run("verifies that the bad request error is thrown", func(t *testing.T) {

			e := echo.New()

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"email": "johndoe@gmail.com", "password": "JohnD0@2123"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.Bind(nil)
			err = as.UserRegister(c)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		t.Run("verifies that the user internal server error is thrown when the user can not be registered", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM users WHERE email = 'johndoe@gmail.com'\) AS email_exists`

			rows := mock.NewRows([]string{"email_exists"}).AddRow(false)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			mock.ExpectExec("INSERT INTO users").
				WithArgs(sqlmock.AnyArg(), "johndoe@gmail.com", sqlmock.AnyArg()).
				WillReturnError(errors.New("internal server error"))

			e := echo.New()

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"email": "johndoe@gmail.com", "password": "JohnD0@2123"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err = as.UserRegister(c)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		})
	})

	t.Run("User Login", func(t *testing.T) {
		t.Run("verifies that the user is logged in successfully", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM users WHERE email = 'example@gmail.com'\) AS email_exists`
			rows := mock.NewRows([]string{"email_exists"}).AddRow(true)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			loginQuery := `SELECT password AS hashPassword, id AS userId FROM users WHERE email = 'example@gmail.com'`
			rowsLogin := mock.NewRows([]string{"hashPassword", "userId"}).AddRow("$2a$10$h1aiikYlNioOFM39E/NmgO8p4QxtVYmQjLqbVJLYeJNUKr/sKU3GG", "ed6caeda-1fa9-442e-a41d-dd2b135cea67")
			mock.ExpectQuery(loginQuery).WillReturnRows(rowsLogin)

			e := echo.New()

			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email": "example@gmail.com", "password": "JohnD0@2123"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err := as.UserLogin(c)
			assert.NoError(t, err)

			var result LoginResponseDto
			err = json.Unmarshal(rec.Body.Bytes(), &result)
			assert.NoError(t, err)

			token, err := CreateJWTToken("example@gmail.com", "ed6caeda-1fa9-442e-a41d-dd2b135cea67")
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, LoginResponseDto{
				Response: constants.USER_LOGIN_SUCCESSFUL,
				Token:    token,
			}, result)
		})

		t.Run("verifies that the bad request error is thrown", func(t *testing.T) {
			e := echo.New()

			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email": "johndoe@gmail.com", "password": "JohnD0@2123"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.Bind(nil)
			err = as.UserLogin(c)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		t.Run("verifies that the user is not logged in when user is not registered", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM users WHERE email = 'example@gmail.com'\) AS email_exists`
			rows := mock.NewRows([]string{"email_exists"}).AddRow(false)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			e := echo.New()

			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email": "example@gmail.com", "password": "JohnD0@2123"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err := as.UserLogin(c)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, rec.Code)
		})

		t.Run("verifies that the internal server error is shown when check email method cannot be called", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM users WHERE email = 'example@gmail.com'\) AS email_exists`
			mock.ExpectQuery(expectedQuery).WillReturnError(sql.ErrNoRows)

			e := echo.New()

			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email": "example@gmail.com", "password": "JohnD0@2123"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err := as.UserLogin(c)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		})

		t.Run("verifies that the internal server error is shown when jwt cannot be created", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM users WHERE email = 'example@gmail.com'\) AS email_exists`
			rows := mock.NewRows([]string{"email_exists"}).AddRow(true)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			loginQuery := `SELECT password AS hashPassword, id AS userId FROM users WHERE email = 'example@gmail.com'`
			rowsLogin := mock.NewRows([]string{"hashPassword", "userId"}).AddRow("$2a$10$h1aiikYlNioOFM39E/NmgO8p4QxtVYmQjLqbVJLYeJNUKr/sKU3GG", "")
			mock.ExpectQuery(loginQuery).WillReturnRows(rowsLogin)

			e := echo.New()

			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email": "example@gmail.com", "password": "JohnD0@2123"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err := as.UserLogin(c)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		})

		t.Run("verifies that the error is thrown when password authorization is failed", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM users WHERE email = 'example@gmail.com'\) AS email_exists`
			rows := mock.NewRows([]string{"email_exists"}).AddRow(true)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			loginQuery := `SELECT password AS hashPassword, id AS userId FROM users WHERE email = 'example@gmail.com'`
			rowsLogin := mock.NewRows([]string{"hashPassword", "userId"}).AddRow("JohnD0@2123", "ed6caeda-1fa9-442e-a41d-dd2b135cea67")
			mock.ExpectQuery(loginQuery).WillReturnRows(rowsLogin)

			e := echo.New()

			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email": "example@gmail.com", "password": "JohnD0@2123"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err := as.UserLogin(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("verifies that the error is thrown when login is failed", func(t *testing.T) {
			expectedQuery := `SELECT EXISTS \(SELECT 1 FROM users WHERE email = 'example@gmail.com'\) AS email_exists`
			rows := mock.NewRows([]string{"email_exists"}).AddRow(true)
			mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

			e := echo.New()

			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email": "example@gmail.com", "password": "JohnD0@2123"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err := as.UserLogin(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		})
	})
}
