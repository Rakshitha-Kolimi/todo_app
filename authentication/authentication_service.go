package authentication

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"todo-project/constants"
)

type Service interface {
	UserRegister(c echo.Context) error
	UserLogin(c echo.Context) error
}

type AuthService struct {
	R AuthRepository
}

func GetUserFromToken(token *jwt.Token) (string, bool) {
	if token != nil {
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return "", ok
		}
		userId, ok := claims["user_id"].(string)
		if !ok {
			return "", ok
		}

		return userId, ok
	}

	return "", false
}

func CreateJWTToken(username, user_id string) (string, error) {
	if username == "" || user_id == "" {
		err := errors.New(constants.INVALID_USERNAME_OR_USER_ID)
		return "", err
	}

	claims := JwtClaims{
		username,
		user_id,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour))},
	}
	jwtSecretKey := os.Getenv("JWT_AUTH_SECRET")

	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	token, err := rawToken.SignedString([]byte(jwtSecretKey))

	if err != nil {
		return "", err
	}

	return token, nil
}

// UserRegister registers a new user
// @Summary Register a new user
// @Description Registers a new user
// @Tags Authentication
// @Accept json
// @Produce plain
// @Param dto body authentication.RegisterDto  true "Email and Password"
// @Success 201 string string constants.USER_REGISTERED_SUCCESSFUL
// @Failure 400 string constants.BAD_REQUEST
// @Failure 403 string constants.EMAIL_ADDRESS_ALREADY_EXISTS
// @Failure 500 string constants.INTERNAL_SERVER_ERROR
// @Router /register [post]
func (s AuthService) UserRegister(c echo.Context) error {
	userRegistrationDto := new(RegisterDto)

	if err := c.Bind(userRegistrationDto); err != nil {
		errMessage := fmt.Sprintf(constants.BAD_REQUEST, err)
		return c.String(http.StatusBadRequest, errMessage)
	}

	email := userRegistrationDto.Email
	password := userRegistrationDto.Password

	isExists, err := s.R.checkEmailExists(email)

	if err != nil {
		errMessage := fmt.Sprintf(constants.CANNOT_CHECK_IF_EMAIL_EXISTS, err)
		return c.String(http.StatusInternalServerError, errMessage)
	}

	if isExists {
		return c.String(http.StatusForbidden, constants.EMAIL_ADDRESS_ALREADY_EXISTS)
	}

	err = s.R.Register(email, password)

	if err != nil {
		errMessage := fmt.Sprintf(constants.INTERNAL_SERVER_ERROR, err)
		return c.String(http.StatusInternalServerError, errMessage)
	}

	return c.String(http.StatusCreated, constants.USER_REGISTERED_SUCCESSFUL)
}

// UserLogin logs in the user
// @Summary Verifies the user and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param dto body authentication.RegisterDto  true "Email and Password"
// @Success 200 {object} authentication.LoginResponseDto
// @Failure 400 string constants.BAD_REQUEST
// @Failure 401 string constants.INVALID_PASSWORD
// @Failure 403 string constants.EMAIL_NOT_REGISTERED
// @Failure 500 string constants.INTERNAL_SERVER_ERROR
// @Router /login [post]
func (s AuthService) UserLogin(c echo.Context) error {
	userRegistrationDto := new(RegisterDto)
	var userId string

	if err := c.Bind(userRegistrationDto); err != nil {
		errMessage := fmt.Sprintf(constants.BAD_REQUEST, err)
		return c.String(http.StatusBadRequest, errMessage)
	}

	email := userRegistrationDto.Email
	password := userRegistrationDto.Password

	isEmailExists, err := s.R.checkEmailExists(email)

	if err != nil {
		errMessage := fmt.Sprintf(constants.INTERNAL_SERVER_ERROR, err)
		return c.String(http.StatusInternalServerError, errMessage)
	}

	if !isEmailExists {
		return c.String(http.StatusForbidden, constants.EMAIL_NOT_REGISTERED)
	}

	hashPassword, userId, err := s.R.Login(email)

	if err != nil {
		return c.String(http.StatusInternalServerError, constants.CANNOT_PROCESS_THE_REQUEST)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))

	if err != nil {
		return c.String(http.StatusUnauthorized, constants.INVALID_PASSWORD)
	}

	token, err := CreateJWTToken(email, userId)
	if err != nil {
		return c.String(http.StatusInternalServerError, constants.CANNOT_PROCESS_THE_REQUEST)
	}

	return c.JSONPretty(http.StatusOK, LoginResponseDto{
		Response: constants.USER_LOGIN_SUCCESSFUL,
		Token:    token,
	}, " ")
}
