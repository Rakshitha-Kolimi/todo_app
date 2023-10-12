package main

import (
	"database/sql"
	"net/http"
	"os"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	"todo-project/api"
	"todo-project/authentication"
	"todo-project/database"
	_ "todo-project/docs"
)

func controller(db *sql.DB, e *echo.Echo) {
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	jwtSecretKey := os.Getenv("JWT_AUTH_SECRET")

	authRepo := authentication.AuthRepository{
		DB: db,
	}

	authService := authentication.AuthService{
		R: authRepo,
	}

	apiRepo := api.ApiRepository{
		DB: db,
	}

	apiService := api.ApiService{
		R: apiRepo,
	}

	e.POST("/register", authService.UserRegister)
	e.POST("/login", authService.UserLogin)

	g := e.Group("/item")

	g.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `[${time_rfc3339}] ${status} ${method} ${host}${path} ${latency_human}` + "\n",
	}))

	g.Use(echojwt.WithConfig(echojwt.Config{
		SigningMethod: "HS512",
		SigningKey:    []byte(jwtSecretKey),
	}))

	g.GET("/main", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to main page")
	})
	g.GET("/:id", apiService.FindById)
	g.POST("/list", apiService.GetAllItems)
	g.POST("/create", apiService.CreateTodoItem)
	g.PUT("/update/:id", apiService.UpdateItemById)
	g.DELETE("/delete/:id", apiService.DeleteById)
	g.PATCH("/complete/:id", apiService.UpdateStatustoCompleted)
}

// @title Todo API
// @version 1.0
// @host localhost:5000
// @BasePath /
// @schemes http https
func main() {
	e := echo.New();
	db := database.ConnectToDb()
	
	controller(db,e)
	defer db.Close()

	e.Logger.Fatal(e.Start(":5000"))
}
