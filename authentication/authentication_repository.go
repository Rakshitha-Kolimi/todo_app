package authentication

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	//checks if the email id already exists
	checkEmailExists(email string) (bool, error)

	//register a new user
	Register(userRegistrationDto RegisterDto) error

	//login using the email and password
	Login(c echo.Context) error
}

type AuthRepository struct {
	DB *sql.DB
  }

func(a *AuthRepository) checkEmailExists(email string) (bool, error) {
	var email_exists bool
	ifEmailExistsQuery := fmt.Sprintf(IfEmailExistsQuery, email)

	err := a.DB.QueryRow(ifEmailExistsQuery).Scan(&email_exists)
	if err != nil {
		return false, err
	}

	return email_exists, nil
}

func (a *AuthRepository) Register(email, password string) error {
	id := uuid.New()

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	registerQuery := CreateUserQuery
	_, err = a.DB.Exec(registerQuery,id,email,string(hashPassword))

	if err != nil {
		return err
	}

	return nil
}

func (a *AuthRepository) Login(email string) (string,string,error) {
	var userId, hashPassword string
	loginQuery := fmt.Sprintf(LoginQuery, email)

	err := a.DB.QueryRow(loginQuery).Scan(&hashPassword, &userId)

	if err != nil {
		return "","",err
	}

    return hashPassword,userId,nil
}